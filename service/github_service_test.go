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

package service

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/stretchr/testify/assert"

	"github.com/jimschubert/changelog/model"
)

func Test_githubService_shouldExcludeViaRepositoryCommit(t *testing.T) {
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
			fields{&model.Config{Exclude: []string{"value"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This value is a commit message")}}},
			true,
		},
		{"should not exclude by non-match in plain text",
			fields{&model.Config{Exclude: []string{"banana"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This value is a commit message")}}},
			false,
		},
		{"excludes by match in simple regular expression",
			fields{&model.Config{Exclude: []string{"\\d+"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a 1234 value commit message")}}},
			true,
		},
		{"should not exclude by non-match in simple regular expression",
			fields{&model.Config{Exclude: []string{"\\d+"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a non-match value commit message")}}},
			false,
		},
		{"excludes by match in complex regular expression",
			fields{&model.Config{Exclude: []string{"(?i)\\d{1,4}\\s(VALUE)"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a 1234 value commit message")}}},
			true,
		},
		{"should not exclude by non-match in complex regular expression",
			fields{&model.Config{Exclude: []string{"(?i)\\d{1,4}\\s(VALUE)"}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a regular commit message")}}},
			false,
		},
		{"should not exclude for nil excludes",
			fields{&model.Config{Exclude: nil}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a regular commit message")}}},
			false,
		},
		{"should not exclude for empty excludes",
			fields{&model.Config{Exclude: []string{}}},
			args{&github.RepositoryCommit{Commit: &github.Commit{Message: p("This is a regular commit message")}}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := githubService{
				config: tt.fields.Config,
			}
			if got := s.shouldExcludeViaRepositoryCommit(tt.args.commit); got != tt.want {
				t.Errorf("shouldExcludeViaRepositoryCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func commitFromJson(t *testing.T, file string) *github.RepositoryCommit {
	t.Helper()
	path := filepath.Join("testdata", file) // relative path
	rawBytes, err := ioutil.ReadFile(path)
	assert.NoError(t, err, "could not find %s", path)
	var commit github.RepositoryCommit
	err = json.Unmarshal(rawBytes, &commit)
	assert.NoError(t, err, "could not unmarshal %s as github.RepositoryCommit", path)

	return &commit
}

func Test_githubService_convertToChangeItem_authorInfo(t *testing.T) {
	background := context.Background()
	type fields struct {
		contextual *Contextual
		config     *model.Config
	}
	type args struct {
		commit *github.RepositoryCommit
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		compare model.ChangeItem
	}{
		{
			name:   "github api doc example commit",
			fields: fields{newContextual(nil), &model.Config{}},
			args:   args{commitFromJson(t, "commit.json")},
			compare: model.ChangeItem{
				AuthorRaw:        github.String("octocat"),
				AuthorURLRaw:     github.String("https://github.com/octocat"),
				CommitMessageRaw: github.String("Fix all the bugs"),
			},
		},
		{
			name:   "issue 1",
			fields: fields{newContextual(nil), &model.Config{}},
			args:   args{commitFromJson(t, "issue_1.json")},
			compare: model.ChangeItem{
				CommitMessageRaw: github.String("Fix all the bugs"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := githubService{
				contextual: tt.fields.contextual,
				config:     tt.fields.config,
			}

			ciChan := make(chan *model.ChangeItem)
			wg := sync.WaitGroup{}
			go func() {
				wg.Add(1)
				s.convertToChangeItem(tt.args.commit, ciChan, &wg, &background)
			}()
			ci := <-ciChan
			assert.NotNil(t, ci)
			assert.Equal(t, tt.compare.Author(), ci.Author())
			assert.Equal(t, tt.compare.AuthorURL(), ci.AuthorURL())
			assert.Equal(t, github.Stringify(tt.compare.CommitMessageRaw), github.Stringify(ci.CommitMessageRaw))
		})
	}
}
