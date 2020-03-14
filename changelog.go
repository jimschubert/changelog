package changelog

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"text/template"
	"time"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"

	"changelog/model"
)

const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
const defaultEnd = "HEAD"
const defaultTemplate = `
## {{.Version}}

{{range .Items -}}
* [{{.CommitHashShort}}]({{.CommitURL}}) {{.Title}} ([{{.Author}}]({{.AuthorURL}}))
{{end}}

<em>For more details, see <a href="{{.CompareURL}}">{{.PreviousVersion}}..{{.Version}}</a></em>
`

// Changelog holds the information required to define the bounds for the changelog
type Changelog struct {
	*model.Config
	From string
	To   string
}

type data struct {
	Version         string
	PreviousVersion string
	Items           []model.ChangeItem
	DiffURL         string
	PatchURL        string
	CompareURL      string
}

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

	compareContext, compareCancel := newContext(context.Background())
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
			c.convertToChangeItem(&commit, ciChan, &wg)
		}(commit)
	}

	go wait(doneChan, &wg)

	all := make([]model.ChangeItem, 0)
	for {
		select {
		case e := <-errorChan:
			return e
		case ci := <-ciChan:
			all = append(all, *ci)
		case <-doneChan:
			return c.writeChangelog(all, comparison, writer)
		}
	}
}

func wait(ch chan struct{}, wg *sync.WaitGroup) {
	wg.Wait()
	ch <- struct{}{}
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

	d := &data{
		PreviousVersion: c.From,
		Version:         c.To,
		Items:           all,
		CompareURL:      compareURL,
		DiffURL:         diffURL,
		PatchURL:        patchURL,
	}

	var tpl = defaultTemplate
	if c.Config.Template != nil {
		b, templateErr := ioutil.ReadFile(*c.Config.Template)
		if templateErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Unable to load template. Using default.")
		} else {
			tpl = string(b)
		}
	}

	tmpl, err := template.New("test").Parse(tpl)
	if err != nil {
		return err
	}

	_ = tmpl.Execute(writer, d)
	return nil
}

func (c *Changelog) convertToChangeItem(commit *github.RepositoryCommit, ch chan *model.ChangeItem, wg *sync.WaitGroup) {
	defer wg.Done()

	var t *time.Time
	if commit.GetCommit() != nil && (*commit.GetCommit()).GetAuthor() != nil && (*(*commit.GetCommit()).GetAuthor()).Date != nil {
		t = (*(*commit.GetCommit()).GetAuthor()).Date
	}

	// TODO: Groupings
	// TODO: Pull URL/Boolean
	// TODO: Excludes
	ci := &model.ChangeItem{
		AuthorRaw:        commit.Author.Login,
		AuthorURLRaw:     commit.Author.URL,
		CommitMessageRaw: commit.Commit.Message,
		DateRaw:          t,
		IsPullRaw:        nil,
		PullURLRaw:       nil,
		CommitHashRaw:    commit.SHA,
		CommitURLRaw:     commit.HTMLURL,
		GroupRaw:         nil,
	}

	ch <- ci
}

type CommitDescendingSorter []model.ChangeItem

func (a CommitDescendingSorter) Len() int           { return len(a) }
func (a CommitDescendingSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitDescendingSorter) Less(i, j int) bool { return a[i].Date().Unix() > a[j].Date().Unix() }

type CommitAscendingSorter []model.ChangeItem

func (a CommitAscendingSorter) Len() int           { return len(a) }
func (a CommitAscendingSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitAscendingSorter) Less(i, j int) bool { return a[i].Date().Unix() < a[j].Date().Unix() }
