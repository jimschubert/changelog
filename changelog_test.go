// Copyright 2020 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package changelog

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-github/v29/github"

	"changelog/model"
)

func commits(items ...model.ChangeItem) *[]model.ChangeItem {
	arr := make([]model.ChangeItem, 0)
	if len(items) > 0 {
		arr = append(arr, items...)
	}
	return &arr
}

func TestCommitDescendingSorter(t *testing.T) {
	present := time.Time{}
	past := present.Add(time.Hour * -5)
	future := present.Add(time.Hour * 12)
	type expect struct {
		first  int64
		second int64
	}
	tests := []struct {
		name string
		a    *[]model.ChangeItem
		args expect
	}{
		{"descending from out of order (past, present)",
			commits(
				model.ChangeItem{DateRaw: &past},
				model.ChangeItem{DateRaw: &present}),
			expect{first: present.Unix(), second: past.Unix()},
		},
		{"descending from in order (future, present)",
			commits(
				model.ChangeItem{DateRaw: &future},
				model.ChangeItem{DateRaw: &present}),
			expect{first: future.Unix(), second: present.Unix()},
		},
		{"descending from same (present, present)",
			commits(
				model.ChangeItem{DateRaw: &present},
				model.ChangeItem{DateRaw: &present}),
			expect{first: present.Unix(), second: present.Unix()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(CommitDescendingSorter(*tt.a))
			if (*tt.a)[0].Date().Unix() != tt.args.first && (*tt.a)[1].Date().Unix() != tt.args.second {
				t.Errorf("CommitDescendingSorter failed to sort %d > %d", tt.args.first, tt.args.second)
			}
		})
	}
}

func TestCommitAscendingSorter(t *testing.T) {
	present := time.Time{}
	past := present.Add(time.Hour * -5)
	future := present.Add(time.Hour * 12)
	type expect struct {
		first  int64
		second int64
	}
	tests := []struct {
		name string
		a    *[]model.ChangeItem
		args expect
	}{
		{"ascending from out of order (past, present)",
			commits(
				model.ChangeItem{DateRaw: &past},
				model.ChangeItem{DateRaw: &present}),
			expect{first: past.Unix(), second: present.Unix()},
		},
		{"ascending from in order (future, present)",
			commits(
				model.ChangeItem{DateRaw: &future},
				model.ChangeItem{DateRaw: &present}),
			expect{first: present.Unix(), second: future.Unix()},
		},
		{"ascending from same (present, present)",
			commits(
				model.ChangeItem{DateRaw: &present},
				model.ChangeItem{DateRaw: &present}),
			expect{first: present.Unix(), second: present.Unix()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(CommitAscendingSorter(*tt.a))
			if (*tt.a)[0].Date().Unix() != tt.args.first && (*tt.a)[1].Date().Unix() != tt.args.second {
				t.Errorf("CommitAscendingSorter failed to sort %d < %d", tt.args.first, tt.args.second)
			}
		})
	}
}

// func TestChangelog_Generate(t *testing.T) {
// 	type fields struct {
// 		Config *model.Config
// 		From   string
// 		To     string
// 	}
// 	tests := []struct {
// 		name       string
// 		fields     fields
// 		wantWriter string
// 		wantErr    bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &Changelog{
// 				Config: tt.fields.Config,
// 				From:   tt.fields.From,
// 				To:     tt.fields.To,
// 			}
// 			writer := &bytes.Buffer{}
// 			err := c.Generate(writer)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
// 				t.Errorf("Generate() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
// 			}
// 		})
// 	}
// }

// func TestChangelog_convertToChangeItem(t *testing.T) {
// 	type fields struct {
// 		Config *model.Config
// 		From   string
// 		To     string
// 	}
// 	type args struct {
// 		commit *github.RepositoryCommit
// 		ch     chan *model.ChangeItem
// 		wg     *sync.WaitGroup
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &Changelog{
// 				Config: tt.fields.Config,
// 				From:   tt.fields.From,
// 				To:     tt.fields.To,
// 			}
// 		})
// 	}
// }

func TestChangelog_writeChangelog(t *testing.T) {
	p := func(s string) *string { return &s }
	gp := func(arr []model.Grouping) *[]model.Grouping { return &arr }
	fromTimestamp := func(ts int64) *time.Time {
		t := time.Unix(ts, 0)
		return &t
	}

	compareUrl := func(c *model.Config, from string, to string) string {
		u := fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s",
			c.Owner,
			c.Repo,
			from,
			to)
		s := fmt.Sprintf("<em>For more details, see <a href=\"%s\">%s..%s</a></em>\n",
			u,
			from,
			to)
		return s
	}

	ci := func(ci *model.ChangeItem) string {
		pullPart := ""
		if ci.IsPull() {
			pullPart = fmt.Sprintf("[contributed](%s) by ", ci.PullURL())
		}
		li := fmt.Sprintf("* [%s](%s) %s (%s[%s](%s))\n",
			ci.CommitHashShort(),
			ci.CommitURL(),
			ci.Title(),
			pullPart,
			ci.Author(),
			ci.AuthorURL(),
		)
		return li
	}

	withGroup := func(group string, ci *model.ChangeItem) model.ChangeItem {
		date := ci.Date()
		return model.ChangeItem{
			AuthorRaw:        ci.AuthorRaw,
			AuthorURLRaw:     ci.AuthorURLRaw,
			CommitMessageRaw: ci.CommitMessageRaw,
			DateRaw:          &date,
			IsPullRaw:        ci.IsPullRaw,
			PullURLRaw:       ci.PullURLRaw,
			CommitHashRaw:    ci.CommitHashRaw,
			CommitURLRaw:     ci.CommitURLRaw,
			GroupRaw:         &group,
		}
	}

	// git log --format="%ct %H %an %s"
	flatConfig := &model.Config{
		Owner:         "jimschubert",
		Repo:          "changelog",
		SortDirection: model.Descending.Ptr(),
		Groupings:     nil,
	}

	groupedConfig := &model.Config{
		Owner:         "jimschubert",
		Repo:          "changelog",
		SortDirection: model.Descending.Ptr(),
		Groupings: gp([]model.Grouping{
			{Name: "Features", Patterns: []string{"(?i)\badd\b"}},
			{Name: "Other", Patterns: []string{".?"}},
		}),
	}

	expectedFlat := func(c *model.Config, from string, to string, items ...model.ChangeItem) string {
		result := &bytes.Buffer{}
		result.WriteString(fmt.Sprintf("## %s\n\n", to))
		for _, item := range items {
			result.WriteString(ci(&item))
		}
		result.WriteString("\n")
		result.WriteString(compareUrl(c, from, to))
		return result.String()
	}

	expectedGrouped := func(c *model.Config, from string, to string, items map[string][]model.ChangeItem) string {
		result := &bytes.Buffer{}
		result.WriteString(fmt.Sprintf("## %s\n", to))
		for key, changeItems := range items {
			result.WriteString("\n")
			result.WriteString("### ")
			result.WriteString(key)
			result.WriteString("\n\n")
			for _, item := range changeItems {
				result.WriteString(ci(&item))
			}
		}

		result.WriteString("\n")
		result.WriteString(compareUrl(c, from, to))
		return result.String()
	}

	// note; initialize ChangeItem without Group here. call setGroup(model.ChangeItem) for group expectations.
	first := model.ChangeItem{
		AuthorRaw:        p("jimschubert"),
		AuthorURLRaw:     p("https://github.com/jimschubert"),
		CommitMessageRaw: p("Initial commit"),
		CommitHashRaw:    p("ae494dca96571b5cf8cd6ad8c9fccf86d8455982"),
		CommitURLRaw:     p("https://github.com/jimschubert/changelog/commit/ae494dca96571b5cf8cd6ad8c9fccf86d8455982"),
		DateRaw:          fromTimestamp(1583008420),
	}
	second := model.ChangeItem{
		AuthorRaw:        p("jimschubert"),
		AuthorURLRaw:     p("https://github.com/jimschubert"),
		CommitMessageRaw: p("Add some placeholder command line args"),
		CommitHashRaw:    p("d707829d23b58326182c3c17fb5f52d275feda6b"),
		CommitURLRaw:     p("https://github.com/jimschubert/changelog/commit/d707829d23b58326182c3c17fb5f52d275feda6b"),
		DateRaw:          fromTimestamp(1583008987),
	}

	type fields struct {
		Config *model.Config
		From   string
		To     string
	}
	type args struct {
		all        []model.ChangeItem
		comparison *github.CommitsComparison
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantWriter string
		wantErr    bool
	}{
		{
			"single item flat output",
			fields{
				Config: flatConfig,
				From:   "v0.0.0",
				To:     "v0.0.1",
			},
			args{
				all: []model.ChangeItem{first},
				comparison: &github.CommitsComparison{
					HTMLURL: p("https://github.com/jimschubert/changelog/compare/v0.0.0...v0.0.1"),
				},
			},
			expectedFlat(flatConfig, "v0.0.0", "v0.0.1", first),
			false,
		},
		{
			"single item grouped output",
			fields{
				Config: groupedConfig,
				From:   "v0.0.0",
				To:     "v0.0.1",
			},
			args{
				all: []model.ChangeItem{withGroup("Other", &first), withGroup("Features", &second)},
				comparison: &github.CommitsComparison{
					HTMLURL: p("https://github.com/jimschubert/changelog/compare/v0.0.0...v0.0.1"),
				},
			},
			expectedGrouped(groupedConfig, "v0.0.0", "v0.0.1", map[string][]model.ChangeItem{
				"Features": {withGroup("Features", &second)},
				"Other":    {withGroup("Other", &first)},
			}),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Changelog{
				Config: tt.fields.Config,
				From:   tt.fields.From,
				To:     tt.fields.To,
			}
			writer := &bytes.Buffer{}
			err := c.writeChangelog(tt.args.all, writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeChangelog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("writeChangelog() gotWriter = '''%v''', want '''%v'''", gotWriter, tt.wantWriter)
			}
		})
	}
}
