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

package service

import (
	"testing"

	"github.com/google/go-github/v29/github"

	"changelog/model"
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
			s := githubService{
				config: tt.fields.Config,
			}
			if got := s.shouldExcludeViaRepositoryCommit(tt.args.commit); got != tt.want {
				t.Errorf("shouldExcludeViaRepositoryCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}
