// Copyright 2020-2026 Jim Schubert
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

package model

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/stretchr/testify/assert"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func createTempConfig(t *testing.T, data string, extension string) (fileLocation string, cleanup func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "Config_test")
	if err != nil {
		t.Fatal(err)
	}
	r := rand.Int()
	testHash := hash(t.Name())
	testFile := fmt.Sprintf("file-%d-%d.%s", r, testHash, extension)
	filePath := filepath.Join(tempDir, testFile)
	if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	return filePath, func() { _ = os.RemoveAll(filePath) }
}

func stringArray(s ...string) []string {
	arr := make([]string, 0)
	if len(s) > 0 {
		arr = append(arr, s...)
	}
	return arr
}

func groupings(g ...Grouping) []Grouping {
	arr := make([]Grouping, 0)
	if len(g) > 0 {
		arr = append(arr, g...)
	}
	return arr
}

func TestConfig_Load(t *testing.T) {
	bt := true
	pint := func(i int) *int { return &i }
	type fields struct {
		JSONData      string
		ResolveType   *ResolveType
		Owner         string
		Repo          string
		Groupings     []Grouping
		Exclude       []string
		Enterprise    *string
		Template      *string
		SortDirection *SortDirection
		TempFileExt   *string
		PreferLocal   *bool
		MaxCommits    *int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"Loads valid empty json", fields{JSONData: "{}"}, false},
		{"Loads valid json resolve-only", fields{JSONData: `{"resolve": "commits"}`, ResolveType: Commits.Ptr()}, false},
		{"Fail on valid json with invalid data type resolve-only", fields{JSONData: `{"resolve": 1.0}`, ResolveType: ResolveType(0).Ptr()}, true}, // note that 1 would resolve since enum is an int
		{"Loads valid json owner-only", fields{JSONData: `{"owner": "jimschubert"}`, Owner: "jimschubert"}, false},
		{"Fail on valid json with invalid data type owner-only", fields{JSONData: `{"owner": []}`}, true},
		{"Loads valid json repo-only", fields{JSONData: `{"repo": "changelog"}`, Repo: "changelog"}, false},
		{"Fail on valid json with invalid data type repo-only", fields{JSONData: `{"repo": []}`}, true},
		{"Loads valid json groupings-only", fields{JSONData: `{"groupings":[{"name":"g","patterns":[]}]}`, Groupings: groupings(Grouping{Name: "g", Patterns: make([]string, 0)})}, false},
		{"Loads valid json exclude-only", fields{JSONData: `{"exclude": []}`, Exclude: stringArray()}, false},
		{"Loads valid json enterprise-only", fields{JSONData: `{"enterprise": "https://ghe.example.com"}`, Enterprise: p("https://ghe.example.com")}, false},
		{"Loads valid json template-only", fields{JSONData: `{"template": "/path/to/template"}`, Template: p("/path/to/template")}, false},
		{"Loads valid json ascending sort-only", fields{JSONData: `{"sort": "asc"}`, SortDirection: Ascending.Ptr()}, false},
		{"Loads valid json descending sort-only", fields{JSONData: `{"sort": "desc"}`, SortDirection: Descending.Ptr()}, false},
		{"Loads valid config_full.json",
			fields{
				JSONData:      string(helperTestData(t, "config_full.json")),
				ResolveType:   Commits.Ptr(),
				Owner:         "jimschubert",
				Repo:          "ossify",
				Groupings:     groupings(Grouping{Name: "feature", Patterns: []string{}}, Grouping{Name: "bug", Patterns: []string{}}),
				Exclude:       stringArray("wip", "help wanted"),
				Enterprise:    p("https://ghe.example.com"),
				Template:      p("/path/to/template"),
				SortDirection: Ascending.Ptr(),
				PreferLocal:   &bt,
				MaxCommits:    pint(150),
			}, false},
		{"Loads valid config_full.yaml",
			fields{
				JSONData:      string(helperTestData(t, "config_full.yaml")),
				ResolveType:   Commits.Ptr(),
				Owner:         "jimschubert",
				Repo:          "ossify",
				Groupings:     groupings(Grouping{Name: "feature", Patterns: []string{"^a", "\\bb$"}}, Grouping{Name: "bug", Patterns: []string{"cba", "\\b\\[f\\]\\b"}}),
				Exclude:       stringArray("wip", "help wanted"),
				Enterprise:    p("https://ghe.example.com"),
				Template:      p("/path/to/template"),
				SortDirection: Ascending.Ptr(),
				TempFileExt:   p("yaml"),
				PreferLocal:   &bt,
				MaxCommits:    pint(199),
			}, false},
		{"Fails on invalid json",
			fields{
				JSONData: `{"resolve":"commits":"owner":"jimschubert"}`,
			}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tmpExt string
			if tt.fields.TempFileExt != nil {
				tmpExt = *tt.fields.TempFileExt
			} else {
				tmpExt = "json"
			}
			jsonLocation, cleanup := createTempConfig(t, tt.fields.JSONData, tmpExt)
			defer cleanup()

			c := &Config{}
			err := c.Load(jsonLocation)
			if tt.wantErr {
				assert.Error(t, err, "Load() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.NoError(t, err)

				if tt.fields.ResolveType == nil {
					assert.Nil(t, c.ResolveType)
				} else {
					assert.Equal(t, tt.fields.ResolveType, c.ResolveType)
				}
				assert.Equal(t, tt.fields.Owner, c.Owner)
				assert.Equal(t, tt.fields.Repo, c.Repo)

				// Compare groupings by name and patterns only (ignore compiled field)
				assert.Equal(t, len(tt.fields.Groupings), len(c.Groupings))
				for i := range tt.fields.Groupings {
					assert.Equal(t, tt.fields.Groupings[i].Name, c.Groupings[i].Name)
					assert.Equal(t, tt.fields.Groupings[i].Patterns, c.Groupings[i].Patterns)
				}

				assert.Equal(t, tt.fields.Exclude, c.Exclude)
				assert.Equal(t, tt.fields.Enterprise, c.Enterprise)
				assert.Equal(t, tt.fields.Template, c.Template)
				assert.Equal(t, tt.fields.PreferLocal, c.PreferLocal)
				assert.Equal(t, tt.fields.MaxCommits, c.MaxCommits)
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	type fields struct {
		ResolveType   *ResolveType
		Owner         string
		Repo          string
		Groupings     []Grouping
		Exclude       []string
		Enterprise    *string
		Template      *string
		SortDirection *SortDirection
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"outputs string",
			fields{
				ResolveType:   Commits.Ptr(),
				Owner:         "jimschubert",
				Repo:          "ossify",
				Groupings:     groupings(Grouping{Name: "feature", Patterns: []string{}}, Grouping{Name: "bug", Patterns: []string{}}),
				Exclude:       stringArray("wip", "help wanted"),
				Enterprise:    p("https://ghe.example.com"),
				Template:      p("/path/to/template"),
				SortDirection: Ascending.Ptr(),
			}, `Config: { ResolveType: commits Owner: jimschubert Repo: ossify Groupings: [{feature []} {bug []}] Exclude: [wip help wanted] Enterprise: https://ghe.example.com Template: /path/to/template Sort: asc }`},

		{"outputs string for nil properties",
			fields{}, `Config: { ResolveType: <nil> Owner:  Repo:  Groupings: [] Exclude: [] Enterprise:  Template:  Sort:  }`},
	}
	// Config: {ResolveType: commits Owner: jimschubert Repo: ossify Groupings: &[feature bug] Exclude: &[wip help wanted] Enterprise: 0xc00003c5b0 Template: 0xc00003c5c0}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ResolveType:   tt.fields.ResolveType,
				Owner:         tt.fields.Owner,
				Repo:          tt.fields.Repo,
				Groupings:     tt.fields.Groupings,
				Exclude:       tt.fields.Exclude,
				Enterprise:    tt.fields.Enterprise,
				Template:      tt.fields.Template,
				SortDirection: tt.fields.SortDirection,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadOrNewConfig(t *testing.T) {
	type args struct {
		path  *string
		owner string
		repo  string
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadOrNewConfig(tt.args.path, tt.args.owner, tt.args.repo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadOrNewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_FindGroup(t *testing.T) {
	p := func(s string) *string {
		return &s
	}
	g := func(grouping ...Grouping) []Grouping {
		arr := make([]Grouping, 0)
		if len(grouping) > 0 {
			arr = append(arr, grouping...)
		}
		return arr
	}
	type fields struct {
		Config *Config
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
			fields{&Config{Groupings: nil}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This value is a commit message")}}},
			nil,
		},
		{"should result in found group when grouping has single option",
			fields{
				&Config{Groupings: g(Grouping{Name: "First", Patterns: []string{"^docs:"}})},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("docs: This value is a commit message")}}},
			p("First"),
		},
		{"should result in first found group when grouping has multiple options (index 0)",
			fields{
				&Config{Groupings: g(
					Grouping{Name: "First", Patterns: []string{"^docs:"}},
					Grouping{Name: "Second", Patterns: []string{"^second:"}},
				)},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("docs: This value is a commit message")}}},
			p("First"),
		},
		{"should result in first found group when grouping has multiple options (index > 0)",
			fields{
				&Config{Groupings: g(
					Grouping{Name: "First", Patterns: []string{"^docs:"}},
					Grouping{Name: "Second", Patterns: []string{"^second:"}},
				)},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("second: This value is a commit message")}}},
			p("Second"),
		},
		{"should support plain-text grouping",
			fields{
				&Config{
					Groupings: g(
						Grouping{Name: "First", Patterns: []string{"^docs:"}},
						Grouping{Name: "Second", Patterns: []string{"plain text"}},
						Grouping{Name: "Second", Patterns: []string{"^second:"}},
					),
				},
			},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("second: This value is a plain text commit message")}}},
			p("Second"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.Config
			if got := c.FindGroup(tt.args.commit.GetCommit().GetMessage()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
