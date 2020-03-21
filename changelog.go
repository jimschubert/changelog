package changelog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/google/go-github/v29/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"changelog/model"
)

const emptyTree = "master~1"
const defaultEnd = "master"
const defaultTemplate = `
## {{.Version}}

{{range .Items -}}
* [{{.CommitHashShort}}]({{.CommitURL}}) {{.Title}} ({{if .IsPull}}[contributed]({{.PullURL}}) by {{end}}[{{.Author}}]({{.AuthorURL}}))
{{end}}

<em>For more details, see <a href="{{.CompareURL}}">{{.PreviousVersion}}..{{.Version}}</a></em>
`

// Changelog holds the information required to define the bounds for the changelog
type Changelog struct {
	*model.Config
	From string
	To   string
}

type clientContext struct {}

func newContext(c context.Context) (context.Context, context.CancelFunc) {
	timeout, cancel := context.WithTimeout(c, 10*time.Second)
	return timeout, cancel
}

// Generate will format a changelog, writing to the supplied writer
func (c *Changelog) Generate(writer io.Writer) error {
	ctx := context.Background()
	token, found := os.LookupEnv("GITHUB_TOKEN")
	if !found {
		fmt.Println("Environment variable GITHUB_TOKEN not found.")
		os.Exit(1)
	}

	if len(c.From) == 0 {
		c.From = emptyTree
	}
	if len(c.To) == 0 {
		c.To = defaultEnd
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	var client *github.Client
	if c.Config.Enterprise != nil && len(*c.Config.Enterprise) > 0 {
		cl, e := github.NewEnterpriseClient(*c.Config.Enterprise, *c.Config.Enterprise, tc)
		if e != nil {
			return e
		}
		client = cl
	} else {
		client = github.NewClient(tc)
	}

	clientCtxt := context.WithValue(ctx, clientContext{}, client)

	compareContext, compareCancel := newContext(clientCtxt)
	defer compareCancel()

	// TODO: Fail if comparison is behind (example v4.0.0..v3.0.0)?
	comparison, _, compareError := client.Repositories.CompareCommits(compareContext, c.Config.Owner, c.Config.Repo, c.From, c.To)

	if compareError != nil {
		return compareError
	}

	doneChan := make(chan struct{})
	errorChan := make(chan error)
	ciChan := make(chan *model.ChangeItem)

	wg := sync.WaitGroup{}

	for _, commit := range (*comparison).Commits {
		wg.Add(1)
		go func(commit github.RepositoryCommit) {
			c.convertToChangeItem(&commit, ciChan, &wg, &clientCtxt)
		}(commit)
	}

	go wait(doneChan, &wg)

	all := make([]model.ChangeItem, 0)
	for {
		select {
		case e := <-errorChan:
			return e
		case ci := <-ciChan:
			if ci != nil {
				all = append(all, *ci)
			}
		case <-doneChan:
			return c.writeChangelog(all, comparison, writer)
		}
	}
}

func wait(ch chan struct{}, wg *sync.WaitGroup) {
	wg.Wait()
	ch <- struct{}{}
}

func (c *Changelog) applyPullPropertiesChangeItem(ci *model.ChangeItem) {
	re := regexp.MustCompile(`.+?#(\d+).+?`)
	title := ci.Title()
	match := re.FindStringSubmatch(title)
	if match != nil && len(match) > 0 {
		isPull := true
		ci.IsPullRaw = &isPull
		baseUrl := ci.CommitURL()
		idx := strings.LastIndex(baseUrl, "commit")
		if idx > 0 {
			var buffer bytes.Buffer
			buffer.WriteString(baseUrl[0:idx])
			buffer.WriteString("pull/")
			buffer.WriteString(match[1])
			result := buffer.String()
			ci.PullURLRaw = &result

			log.WithFields(log.Fields{
				"baseURL": baseUrl,
				"commitIdx": idx,
				"pullURL": ci.PullURL(),
				"isPull": ci.IsPull(),
			}).Debug("applyPullPropertiesChangeItem")
		}
	}
}

func (c *Changelog) writeChangelog(all []model.ChangeItem, comparison *github.CommitsComparison, writer io.Writer) error {
	compareURL := comparison.GetHTMLURL()
	diffURL := comparison.GetDiffURL()
	patchURL := comparison.GetPatchURL()

	switch *c.Config.SortDirection {
	case model.Ascending:
		sort.Sort(CommitAscendingSorter(all))
	case model.Descending:
		sort.Sort(CommitDescendingSorter(all))
	}

	grouped := make(map[string][]model.ChangeItem)
	for _, item := range all {
		grouped[item.Group()] = append(grouped[item.Group()], item)
	}

	d := &model.TemplateData{
		PreviousVersion: c.From,
		Version:         c.To,
		Items:           all,
		CompareURL:      compareURL,
		DiffURL:         diffURL,
		PatchURL:        patchURL,
		Grouped:         grouped,
	}

	var tpl = defaultTemplate
	if c.Config.Template != nil {
		b, templateErr := ioutil.ReadFile(*c.Config.Template)
		if templateErr != nil {
			log.Warn("Unable to load template. Using default.")
		} else {
			log.Debug("Using default template.")
			tpl = string(b)
		}
	}

	tmpl, err := template.New("changelog").Parse(tpl)
	if err != nil {
		return err
	}

	_ = tmpl.Execute(writer, d)
	return nil
}

func (c *Changelog) shouldExcludeViaPullAttributes(pullId int, ctx *context.Context) bool {
	client, ok := (*ctx).Value(clientContext{}).(*github.Client); if !ok {
		log.Warn("Client context not available for checking pull requests")
		return false
	}

	timeout, cancel := newContext(*ctx)
	defer cancel()

	log.Debugf("Checking pull request %d", pullId)
	pr, _, e := client.PullRequests.Get(timeout, c.Config.Owner, c.Config.Repo, pullId)
	if e != nil || pr == nil {
		return false
	}
	if c.shouldExcludeByText(pr.Title) {
		return true
	}
	for _, label := range pr.Labels {
		if c.shouldExcludeByText(label.Name) {
			return true
		}
	}
	return false
}

func (c *Changelog) convertToChangeItem(commit *github.RepositoryCommit, ch chan *model.ChangeItem, wg *sync.WaitGroup, ctx *context.Context) {
	defer wg.Done()
	var isMergeCommit = false
	if commit.GetCommit() != nil && len(commit.GetCommit().Parents) > 1 {
		isMergeCommit = true
	}

	if !isMergeCommit {
		if !c.shouldExclude(commit) {
			excludeByGroup := false
			var t *time.Time
			if commit.GetCommit() != nil && (*commit.GetCommit()).GetAuthor() != nil && (*(*commit.GetCommit()).GetAuthor()).Date != nil {
				t = (*(*commit.GetCommit()).GetAuthor()).Date
			}

			grouping := c.findGroup(commit)
			excludeByGroup = c.shouldExcludeByText(grouping)

			if !excludeByGroup {
				// TODO: Max count?
				ci := &model.ChangeItem{
					AuthorRaw:        commit.Author.Login,
					AuthorURLRaw:     commit.Author.HTMLURL,
					CommitMessageRaw: commit.Commit.Message,
					DateRaw:          t,
					CommitHashRaw:    commit.SHA,
					CommitURLRaw:     commit.HTMLURL,
					GroupRaw:         grouping,
				}
				c.applyPullPropertiesChangeItem(ci)

				if ci.IsPull() {
					pullId, e := ci.PullID(); if e != nil {
						// In the unlikely case that an unexpected pull url is provided by GitHub API, just emit the change item
						ch <- ci
					} else {
						pr, _ := strconv.Atoi(pullId)
						if !c.shouldExcludeViaPullAttributes(pr, ctx) {
							ch <- ci
						}
					}
				} else {
					ch <- ci
				}
			}
		}
	}
}

func (c *Changelog) shouldExclude(commit *github.RepositoryCommit) bool {
	if c.Exclude != nil && len(*c.Exclude) > 0 {
		title := strings.Split(commit.GetCommit().GetMessage(), "\n")[0]
		return c.shouldExcludeByText(&title)
	}

	return false
}

func (c *Changelog) shouldExcludeByText(text *string) bool {
	if text == nil || c.Exclude == nil || len(*c.Exclude) == 0 {
		return false
	}
	for _, pattern := range *c.Exclude {
		re := regexp.MustCompile(pattern)
		if re.Match([]byte(*text)) {
			log.WithFields(log.Fields{"text": *text,"pattern":pattern}).Debug("exclude via pattern")
			return true
		}
	}
	return false
}

func (c *Changelog) findGroup(commit *github.RepositoryCommit) *string {
	var grouping *string
	if c.Groupings != nil {
		title := strings.Split(commit.GetCommit().GetMessage(), "\n")[0]
		for _, g := range *c.Groupings {
			for _, pattern := range g.Patterns {
				re := regexp.MustCompile(pattern)
				if re.Match([]byte(title)) {
					grouping = &g.Name
					log.WithFields(log.Fields{"grouping":*grouping,"title":title}).Debug("found group name for commit")
					return grouping
				}
			}
		}
	}
	return grouping
}

type CommitDescendingSorter []model.ChangeItem

func (a CommitDescendingSorter) Len() int           { return len(a) }
func (a CommitDescendingSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitDescendingSorter) Less(i, j int) bool { return a[i].Date().Unix() > a[j].Date().Unix() }

type CommitAscendingSorter []model.ChangeItem

func (a CommitAscendingSorter) Len() int           { return len(a) }
func (a CommitAscendingSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitAscendingSorter) Less(i, j int) bool { return a[i].Date().Unix() < a[j].Date().Unix() }
