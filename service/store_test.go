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
	"testing"

	"github.com/jimschubert/changelog/model"
)

func Test_applyPullPropertiesChangeItem(t *testing.T) {
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
			applyPullPropertiesChangeItem(tt.args.ci)
			if gotURL := tt.args.ci.PullURL(); gotURL != tt.args.expectURL {
				t.Errorf("applyPullPropertiesChangeItem() PullURL = %v, want = %v", gotURL, tt.args.expectURL)
			}
			if gotIsPR := tt.args.ci.IsPull(); gotIsPR != tt.args.expectIsPR {
				t.Errorf("applyPullPropertiesChangeItem() IsPR = %v, want = %v", gotIsPR, tt.args.expectIsPR)
			}
		})
	}
}
