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
	"reflect"
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

func TestChangelog_applyPullPropertiesChangeItem(t *testing.T) {
	p := func(s string) *string {
		return &s
	}
	type fields struct {
		Config *model.Config
	}
	type args struct {
		ci         *model.ChangeItem
		expectURL  string
		expectIsPR bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"should not apply PR properties when no considered a PR",

			fields{&model.Config{ResolveType: model.Commits.Ptr()}},
			args{
				&model.ChangeItem{CommitMessageRaw: p("Some Commit Message with numbers 12345 and no # symbol preceding")},
				"",
				false,
			},
		},
		{"should apply PR properties when formatted as 'Merge pull request #523 …'",

			fields{&model.Config{ResolveType: model.Commits.Ptr()}},
			args{
				&model.ChangeItem{
					CommitMessageRaw: p("Merge pull request #523 from cli/title-body-web"),
					CommitURLRaw:     p("https://github.com/cli/cli/commit/b5d0b7c640ad897f395a72074a0f4b31787e5826"),
				},
				"https://github.com/cli/cli/pull/523",
				true,
			},
		},
		{"should apply PR properties when formatted as 'Some commit message (#1234)'",

			fields{&model.Config{ResolveType: model.Commits.Ptr()}},
			args{
				&model.ChangeItem{
					CommitMessageRaw: p("Fix Swift4 CI tests (#5540)"),
					CommitURLRaw:     p("https://github.com/OpenAPITools/openapi-generator/commit/728d03b318a3fd4726c93c0f710bb5bedd1f61ab"),
				},
				"https://github.com/OpenAPITools/openapi-generator/pull/5540",
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Changelog{
				Config: tt.fields.Config,
			}
			c.applyPullPropertiesChangeItem(tt.args.ci)
			if gotURL := tt.args.ci.PullURL(); gotURL != tt.args.expectURL {
				t.Errorf("applyPullPropertiesChangeItem() PullURL = %v, want = %v", gotURL, tt.args.expectURL)
			}
			if gotIsPR := tt.args.ci.IsPull(); gotIsPR != tt.args.expectIsPR {
				t.Errorf("applyPullPropertiesChangeItem() IsPR = %v, want = %v", gotIsPR, tt.args.expectIsPR)
			}
		})
	}
}

func TestChangelog_findGroup(t *testing.T) {
	p := func(s string) *string {
		return &s
	}
	g := func(grouping ...model.Grouping) *[]model.Grouping {
		arr := make([]model.Grouping, 0)
		if len(grouping) > 0 {
			arr = append(arr, grouping...)
		}
		return &arr
	}
	type fields struct {
		Config *model.Config
	}
	type args struct {
		commit *github.RepositoryCommit
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *string
	}{
		{"should result in nil group when grouping is nil",
			fields{&model.Config{Groupings: nil}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This value is a commit message")}}},
			nil,
		},
		{"should result in found group when grouping has single option",
			fields{
				&model.Config{Groupings: g(model.Grouping{Name: "First", Patterns: []string{"^docs:"}})},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("docs: This value is a commit message")}}},
			p("First"),
		},
		{"should result in first found group when grouping has multiple options (index 0)",
			fields{
				&model.Config{Groupings: g(
					model.Grouping{Name: "First", Patterns: []string{"^docs:"}},
					model.Grouping{Name: "Second", Patterns: []string{"^second:"}},
				)},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("docs: This value is a commit message")}}},
			p("First"),
		},
		{"should result in first found group when grouping has multiple options (index > 0)",
			fields{
				&model.Config{Groupings: g(
					model.Grouping{Name: "First", Patterns: []string{"^docs:"}},
					model.Grouping{Name: "Second", Patterns: []string{"^second:"}},
				)},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("second: This value is a commit message")}}},
			p("Second"),
		},
		{"should support plain-text grouping",
			fields{
				&model.Config{
					Groupings: g(
						model.Grouping{Name: "First", Patterns: []string{"^docs:"}},
						model.Grouping{Name: "Second", Patterns: []string{"plain text"}},
						model.Grouping{Name: "Second", Patterns: []string{"^second:"}},
					),
				},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("second: This value is a plain text commit message")}}},
			p("Second"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Changelog{
				Config: tt.fields.Config,
			}
			if got := c.findGroup(tt.args.commit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangelog_shouldExclude(t *testing.T) {
	p := func(s string) *string {
		return &s
	}
	type fields struct {
		Config *model.Config
	}
	type args struct {
		commit *github.RepositoryCommit
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"excludes by match in plain text",
			fields{&model.Config{Exclude: &[]string{"value"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This value is a commit message")}}},
			true,
		},
		{"should not exclude by non-match in plain text",
			fields{&model.Config{Exclude: &[]string{"banana"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This value is a commit message")}}},
			false,
		},
		{"excludes by match in simple regular expression",
			fields{&model.Config{Exclude: &[]string{"\\d+"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a 1234 value commit message")}}},
			true,
		},
		{"should not exclude by non-match in simple regular expression",
			fields{&model.Config{Exclude: &[]string{"\\d+"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a non-match value commit message")}}},
			false,
		},
		{"excludes by match in complex regular expression",
			fields{&model.Config{Exclude: &[]string{"(?i)\\d{1,4}\\s(VALUE)"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a 1234 value commit message")}}},
			true,
		},
		{"should not exclude by non-match in complex regular expression",
			fields{&model.Config{Exclude: &[]string{"(?i)\\d{1,4}\\s(VALUE)"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a regular commit message")}}},
			false,
		},
		{"should not exclude for nil excludes",
			fields{&model.Config{Exclude: nil}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a regular commit message")}}},
			false,
		},
		{"should not exclude for empty excludes",
			fields{&model.Config{Exclude: &[]string{}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a regular commit message")}}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Changelog{
				Config: tt.fields.Config,
			}
			if got := c.shouldExclude(tt.args.commit); got != tt.want {
				t.Errorf("shouldExclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangelog_writeChangelog(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Changelog{
				Config: tt.fields.Config,
				From:   tt.fields.From,
				To:     tt.fields.To,
			}
			writer := &bytes.Buffer{}
			err := c.writeChangelog(tt.args.all, tt.args.comparison, writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeChangelog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("writeChangelog() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
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
