// Copyright 2026 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package changelog

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/google/go-github/v29/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/jimschubert/changelog/model"
	"github.com/jimschubert/changelog/service"
)

const emptyTree = "master~1"
const defaultEnd = "master"
const defaultTemplate = `{{define "PullTemplate"}} ({{if .IsPull -}}
{{if .PullURL}}[contributed]({{.PullURL}}){{else}}contributed{{end}} by {{end}}{{if .AuthorURL -}}
[{{.Author}}]({{.AuthorURL}}){{else}}{{.Author}}{{end -}})
{{- end -}}
{{define "CommitTemplate" -}}
{{if .CommitURL}}[{{.CommitHashShort}}]({{.CommitURL}}){{else}}{{.CommitHashShort}}{{end -}}
{{- end -}}
{{define "GroupTemplate" -}}
{{- range .Grouped}}
### {{ .Name }}

{{range .Items -}}
* {{template "CommitTemplate" . }} {{.Title}}{{template "PullTemplate" . }}
{{end -}}
{{end -}}
{{end -}}
{{define "FlatTemplate" -}}
{{range .Items -}}
* {{template "CommitTemplate" . }} {{.Title}}{{template "PullTemplate" . }}
{{end -}}
{{end -}}
{{define "DefaultTemplate" -}}
## {{.Version}}
{{if len .Grouped -}}
{{template "GroupTemplate" . -}}   
{{- else}}
{{template "FlatTemplate" . -}}
{{end}}
<em>For more details, see <a href="{{.CompareURL}}">{{.PreviousVersion}}..{{.Version}}</a></em>
{{end -}}
{{template "DefaultTemplate" . -}}
`

// Changelog holds the information required to define the bounds for the changelog
type Changelog struct {
	*model.Config
	From string
	To   string
}

// Generate will format a changelog, writing to the supplied writer
func (c *Changelog) Generate(writer io.Writer) error {
	ctx := context.Background()
	token, found := os.LookupEnv("GITHUB_TOKEN")
	if !found {
		log.Fatal("Environment variable GITHUB_TOKEN not found.")
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

	doneChan := make(chan struct{})
	errorChan := make(chan error)
	ciChan := make(chan *model.ChangeItem)

	wg := sync.WaitGroup{}

	var target service.Store
	if c.Config.GetPreferLocal() {
		target = service.NewLocalGitService().WithClient(client).WithConfig(c.Config)
	} else {
		target = service.NewGitHubService().WithClient(client).WithConfig(c.Config)
	}
	err := target.Process(&ctx, &wg, ciChan, c.From, c.To)
	if err != nil {
		return err
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
			return c.writeChangelog(all, writer)
		}
	}
}

func (c *Changelog) GetGitURLs() (*model.GitURLs, error) {
	gh := "https://github.com"
	if c.Enterprise != nil {
		gh = strings.TrimSuffix(*c.Enterprise, "/api")
	}
	u, err := url.Parse(gh)
	if err != nil {
		return nil, err
	}
	create := func(op string) string {
		end := fmt.Sprintf("%s...%s%s", c.From, c.To, op)
		u.Path = path.Join(u.Path, c.Owner, c.Repo, "compare", end)
		return u.String()
	}
	return &model.GitURLs{
		CompareURL: create(""),
		DiffURL:    create(".diff"),
		PatchURL:   create(".patch"),
	}, nil
}

func wait(ch chan struct{}, wg *sync.WaitGroup) {
	wg.Wait()
	ch <- struct{}{}
}

func (c *Changelog) writeChangelog(all []model.ChangeItem, writer io.Writer) error {
	var compareURL = ""
	var diffURL = ""
	var patchURL = ""

	u, err := c.GetGitURLs()
	if err != nil {
		log.Warn("Unable to determine urls for compare, diff, and patch.")
	} else {
		compareURL = u.CompareURL
		diffURL = u.DiffURL
		patchURL = u.PatchURL
	}

	switch *c.Config.SortDirection {
	case model.Ascending:
		sort.Sort(CommitAscendingSorter(all))
	case model.Descending:
		sort.Sort(CommitDescendingSorter(all))
	}

	grouped := make(map[string][]model.ChangeItem)
	for _, item := range all {
		g := item.Group()
		if len(g) > 0 {
			grouped[g] = append(grouped[g], item)
		}
	}

	templateGroups := make([]model.TemplateGroup, 0)

	if len(c.Groupings) > 0 {
		for _, grouping := range c.Groupings {
			if grouping.Name != "" {
				if items, ok := grouped[grouping.Name]; ok && len(items) > 0 {
					log.WithFields(log.Fields{
						"name":  grouping.Name,
						"count": len(items),
					}).Debug("found template grouping data")
					templateGroups = append(templateGroups, model.TemplateGroup{
						Name:  grouping.Name,
						Items: items,
					})
				}
			}
		}
	}

	d := &model.TemplateData{
		PreviousVersion: c.From,
		Version:         c.To,
		Items:           all,
		CompareURL:      compareURL,
		DiffURL:         diffURL,
		PatchURL:        patchURL,
		Grouped:         templateGroups,
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

type CommitDescendingSorter []model.ChangeItem

func (a CommitDescendingSorter) Len() int           { return len(a) }
func (a CommitDescendingSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitDescendingSorter) Less(i, j int) bool { return a[i].Date().Unix() > a[j].Date().Unix() }

type CommitAscendingSorter []model.ChangeItem

func (a CommitAscendingSorter) Len() int           { return len(a) }
func (a CommitAscendingSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitAscendingSorter) Less(i, j int) bool { return a[i].Date().Unix() < a[j].Date().Unix() }
