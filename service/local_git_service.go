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
	"context"
	"errors"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v29/github"
	log "github.com/sirupsen/logrus"

	"github.com/jimschubert/changelog/model"
)

var foundError = errors.New("found")

type gitService struct {
	contextual *Contextual
	config     *model.Config
}

func NewLocalGitService() Store {
	service := gitService{}
	return service
}

func (s gitService) WithClient(client *github.Client) Store {
	s.contextual = newContextual(client)
	return s
}

func (s gitService) WithConfig(config *model.Config) Store {
	s.config = config
	return s
}

func (s gitService) GetContextual() *Contextual {
	return s.contextual
}

func (s gitService) Process(parentContext *context.Context, wg *sync.WaitGroup, ciChan chan *model.ChangeItem, from string, to string) error {
	wg.Add(1)
	defer wg.Done()

	contextual := s.contextual

	_, cancel := contextual.CreateContext(parentContext)
	defer cancel()

	dir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Unable to determine current directory for repository.")
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Unable to open current directory as a git repository.")
	}

	//noinspection GoNilness
	fromTag, err := repo.Tag(from)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "from": from}).Fatalf("Unable to find 'from' tag.")
	}

	toTag, err := repo.Tag(to)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "to": to}).Fatalf("Unable to find 'to' tag.")
	}

	commitIter, err := repo.Log(&git.LogOptions{From: fromTag.Hash()})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Unable to retrieve commit information from git repository.")
	}

	var iterated []plumbing.Hash

	//noinspection GoNilness
	err = commitIter.ForEach(func(c *object.Commit) error {
		// If no previous tag is found then from and to are equal
		if fromTag.Hash() == toTag.Hash() {
			return nil
		}
		if c.Hash == toTag.Hash() {
			return foundError
		}
		iterated = append(iterated, c.Hash)
		return nil
	})

	if err != nil && err != foundError {
		return err
	}

	for _, hash := range iterated {
		commit, err := repo.CommitObject(hash)
		if err != nil {
			log.WithFields(log.Fields{"commit": hash}).Fatalf("Failed while process commits.")
		}

		wg.Add(1)
		go func(commit *object.Commit) {
			newContext, newCancel := contextual.CreateContext(parentContext)
			defer newCancel()
			s.convertToChangeItem(commit, ciChan, wg, &newContext)
		}(commit)
	}
	return nil
}

func (s *gitService) convertToChangeItem(commit *object.Commit, ch chan *model.ChangeItem, wg *sync.WaitGroup, ctx *context.Context) {
	defer wg.Done()

	// URL creation is duplicated with GetGitURLs, this could be moved elsewhere to reduce duplication
	gh := "https://github.com"
	if s.config.Enterprise != nil {
		gh = strings.TrimRight(*s.config.Enterprise, "/api")
	}
	u, _ := url.Parse(gh)
	createCommitLocation := func(hash string) string {
		u.Path = path.Join(u.Path, s.config.Owner, s.config.Repo, "commit", hash)
		return u.String()
	}

	var isMergeCommit = commit.NumParents() > 1
	if !isMergeCommit {
		if !s.shouldExcludeViaRepositoryCommit(commit) {
			excludeByGroup := false
			t := &commit.Committer.When
			grouping := (*s.config).FindGroup(commit.Message)
			excludeByGroup = (*s.config).ShouldExcludeByText(grouping)

			if !excludeByGroup {
				hash := commit.Hash.String()
				commitLocation := createCommitLocation(hash)
				ci := &model.ChangeItem{
					AuthorRaw:        &commit.Author.Name,
					CommitMessageRaw: &commit.Message,
					DateRaw:          t,
					CommitHashRaw:    &hash,
					CommitURLRaw:     &commitLocation,
					GroupRaw:         grouping,
				}

				applyPullPropertiesChangeItem(ci)
				if ci.IsPull() {
					pullId, e := ci.PullID()
					if e != nil {
						// In the unlikely case that an unexpected pull url is provided by GitHub API, just emit the change item
						ch <- ci
					} else {
						pr, _ := strconv.Atoi(pullId)
						contextual := s.contextual
						pullRequest, exclude := shouldExcludeViaPullAttributes(pr, contextual, ctx, s.config)
						if !exclude {
							ci.PullURLRaw = pullRequest.HTMLURL
							ci.AuthorURLRaw = pullRequest.GetUser().HTMLURL
							ci.AuthorRaw = pullRequest.GetUser().Login

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

func (s *gitService) shouldExcludeViaRepositoryCommit(commit *object.Commit) bool {
	if s.config == nil {
		return false
	}

	if (*s.config).Exclude != nil && len(*(*s.config).Exclude) > 0 {
		title := strings.Split(commit.Message, "\n")[0]
		return (*s.config).ShouldExcludeByText(&title)
	}

	return false
}
