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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v29/github"

	"github.com/jimschubert/changelog/model"
)

type githubService struct {
	contextual *Contextual
	config     *model.Config
}

// NewGitHubService creates a new Store for accessing commits from the GitHub API
func NewGitHubService() Store {
	service := githubService{}
	return service
}

// WithClient applies a GitHub client to the Store
func (s githubService) WithClient(client *github.Client) Store {
	s.contextual = newContextual(client)
	return s
}

// WithConfig applies a Config instance to the Store
func (s githubService) WithConfig(config *model.Config) Store {
	s.config = config
	return s
}

// GetContextual returns the context wrapper for this service
func (s githubService) GetContextual() *Contextual {
	return s.contextual
}

// Process queries the service for commits, converting to a ChangeItem and sending to the channel
func (s githubService) Process(parentContext *context.Context, wg *sync.WaitGroup, ciChan chan *model.ChangeItem, from string, to string) error {
	contextual := s.contextual

	compareContext, cancel := contextual.CreateContext(parentContext)
	defer cancel()

	client := contextual.GetClient()

	comparison, _, compareError := client.Repositories.CompareCommits(compareContext, (*s.config).Owner, (*s.config).Repo, from, to)
	if compareError != nil {
		return compareError
	}

	maximum := min(len((*comparison).Commits), (*s.config).GetMaxCommits())
	commits := make([]github.RepositoryCommit, maximum)

	copy(commits, (*comparison).Commits)
	for _, commit := range commits {
		wg.Add(1)
		go func(commit github.RepositoryCommit) {
			newContext, newCancel := contextual.CreateContext(parentContext)
			defer newCancel()
			s.convertToChangeItem(&commit, ciChan, wg, &newContext)
		}(commit)
	}

	return nil
}

func (s githubService) convertToChangeItem(commit *github.RepositoryCommit, ch chan *model.ChangeItem, wg *sync.WaitGroup, ctx *context.Context) {
	defer wg.Done()
	var isMergeCommit = false
	if commit.GetCommit() != nil && len(commit.GetCommit().Parents) > 1 {
		isMergeCommit = true
	}

	if !isMergeCommit {
		if !s.shouldExcludeViaRepositoryCommit(commit) {
			excludeByGroup := false
			var t *time.Time
			var authorRaw *string
			var authorUrlRaw *string
			if commit.GetCommit() != nil {
				commitAuthor := commit.GetCommit().GetAuthor()
				d := commitAuthor.GetDate()
				t = &d
			}
			if commit.Author != nil {
				authorRaw = commit.Author.Login
				authorUrlRaw = commit.Author.HTMLURL
			}

			grouping := (*s.config).FindGroup(commit.GetCommit().GetMessage())
			excludeByGroup = (*s.config).ShouldExcludeByText(grouping)

			if !excludeByGroup {
				// TODO: Max count?
				ci := &model.ChangeItem{
					AuthorRaw:        authorRaw,
					AuthorURLRaw:     authorUrlRaw,
					CommitMessageRaw: commit.Commit.Message,
					DateRaw:          t,
					CommitHashRaw:    commit.SHA,
					CommitURLRaw:     commit.HTMLURL,
					GroupRaw:         grouping,
				}

				applyPullPropertiesChangeItem(ci)

				if ci.IsPull() {
					pullId, e := ci.PullID()
					if e != nil {
						// In the unlikely case that an unexpected pull url is provided by GitHub API, just emit the change item
						ch <- ci
					} else {
						// ignoring error here is intentional. if the ID is not parseable (should never happen), just evaluate the rules.
						// the API call to retrieve PR will then also fail and exclude will be false.
						pr, _ := strconv.Atoi(pullId)
						contextual := s.contextual
						_, exclude := shouldExcludeViaPullAttributes(pr, contextual, ctx, s.config)
						if !exclude {
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

func (s githubService) shouldExcludeViaRepositoryCommit(commit *github.RepositoryCommit) bool {
	if s.config == nil {
		return false
	}

	if (*s.config).Exclude != nil && len(*(*s.config).Exclude) > 0 {
		title := strings.Split(commit.GetCommit().GetMessage(), "\n")[0]
		return (*s.config).ShouldExcludeByText(&title)
	}

	return false
}
