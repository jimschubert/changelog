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
	"errors"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v29/github"
	log "github.com/sirupsen/logrus"

	"github.com/jimschubert/changelog/model"
)

type gitService struct {
	contextual *Contextual
	config     *model.Config
}

func NewLocalGitService() Store {
	service := &gitService{}
	return service
}

func (s *gitService) WithClient(client *github.Client) Store {
	s.contextual = newContextual(client)
	return s
}

func (s *gitService) WithConfig(config *model.Config) Store {
	s.config = config
	return s
}

func (s *gitService) GetContextual() *Contextual {
	return s.contextual
}

//goland:noinspection ALL
func (s *gitService) Process(parentContext *context.Context, wg *sync.WaitGroup, ciChan chan *model.ChangeItem, from string, to string) error {
	wg.Add(1)
	defer wg.Done()

	contextual := s.contextual

	_, cancel := contextual.CreateContext(parentContext)
	defer cancel()

	dir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to determine current directory for repository.")
		return err
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to open current directory as a git repository.")
		return err
	}

	//noinspection GoNilness
	fromTag, err := repo.Tag(from)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "from": from}).Error("Unable to find 'from' tag.")
		return err
	}

	toTag, err := repo.Tag(to)
	if toTag == nil || err != nil {
		log.WithFields(log.Fields{"error": err, "to": to}).Error("Unable to find 'to' tag.")
		return err
	}

	startCommit, err := repo.CommitObject(toTag.Hash())
	if err != nil {
		return err
	}

	// NewCommitIterBSF returns a CommitIter that walks the commit history,
	// starting at the given commit and visiting its parents in pre-order.
	err = object.NewCommitIterBSF(startCommit, nil, nil).ForEach(func(commit *object.Commit) error {
		if commit.Hash == fromTag.Hash() {
			return io.EOF
		}
		wg.Add(1)
		go func(commit *object.Commit) {
			newContext, newCancel := contextual.CreateContext(parentContext)
			defer newCancel()
			s.convertToChangeItem(commit, ciChan, wg, &newContext)
		}(commit)
		return nil
	})

	if err != nil && !errors.Is(err, io.EOF) {
		log.WithFields(log.Fields{
			"from": fromTag.Hash().String(),
			"to":   toTag.Hash().String(),
		}).Error("Failed while processing commits.")
		return err
	}

	return nil
}

func (s *gitService) convertToChangeItem(commit *object.Commit, ch chan *model.ChangeItem, wg *sync.WaitGroup, ctx *context.Context) {
	defer wg.Done()

	// Early returns to reduce nesting
	if commit.NumParents() > 1 {
		return // Skip merge commits
	}

	if s.shouldExcludeViaRepositoryCommit(commit) {
		return
	}

	grouping := s.config.FindGroup(commit.Message)
	if s.config.ShouldExcludeByText(grouping) {
		return
	}

	// URL creation is duplicated with GetGitURLs, this could be moved elsewhere to reduce duplication
	gh := "https://github.com"
	if s.config.Enterprise != nil {
		gh = strings.TrimSuffix(*s.config.Enterprise, "/api")
	}
	baseURL, err := url.Parse(gh)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "url": gh}).Warn("Unable to parse base URL for commit links")
		return
	}

	hash := commit.Hash.String()
	// Create a copy to avoid mutating the shared baseURL
	u := *baseURL
	u.Path = path.Join(u.Path, s.config.Owner, s.config.Repo, "commit", hash)
	commitLocation := u.String()
	t := &commit.Committer.When

	ci := &model.ChangeItem{
		AuthorRaw:        &commit.Author.Name,
		CommitMessageRaw: &commit.Message,
		DateRaw:          t,
		CommitHashRaw:    &hash,
		CommitURLRaw:     &commitLocation,
		GroupRaw:         grouping,
	}

	applyPullPropertiesChangeItem(ci)

	if !ci.IsPull() {
		ch <- ci
		return
	}

	pullId, e := ci.PullID()
	if e != nil {
		// In the unlikely case that an unexpected pull url is provided by GitHub API, just emit the change item
		ch <- ci
		return
	}

	pr, _ := strconv.Atoi(pullId)
	pullRequest, exclude := shouldExcludeViaPullAttributes(pr, s.contextual, ctx, s.config)
	if exclude {
		return
	}

	ci.PullURLRaw = pullRequest.HTMLURL
	ci.AuthorURLRaw = pullRequest.GetUser().HTMLURL
	ci.AuthorRaw = pullRequest.GetUser().Login
	ch <- ci
}

func (s *gitService) shouldExcludeViaRepositoryCommit(commit *object.Commit) bool {
	if s.config == nil {
		return false
	}

	if len(s.config.Exclude) > 0 {
		title, _, _ := strings.Cut(commit.Message, "\n")
		return s.config.ShouldExcludeByText(&title)
	}

	return false
}
