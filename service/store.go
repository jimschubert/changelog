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
	"bytes"
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/google/go-github/v29/github"
	log "github.com/sirupsen/logrus"

	"github.com/jimschubert/changelog/model"
)

// Store defines the functional interface for accessing a store of Git commits
type Store interface {
	// WithClient applies a client to the store
	WithClient(client *github.Client) Store
	// WithConfig applies a config to to the store
	WithConfig(config *model.Config) Store
	// GetContextual returns the context wrapper associated with this store
	GetContextual() *Contextual
	// Process queries the store and converts commits to a ChangeItem before sending to the channel
	Process(parentContext *context.Context, wg *sync.WaitGroup, ciChan chan *model.ChangeItem, from string, to string) error
}

func applyPullPropertiesChangeItem(ci *model.ChangeItem) {
	re := regexp.MustCompile(`.+?#(\d+).+?`)
	title := ci.Title()
	match := re.FindStringSubmatch(title)
	if match != nil && len(match) > 0 {
		isPull := true
		ci.IsPullRaw = &isPull
		baseUrl := ci.CommitURL()
		idx := strings.LastIndex(baseUrl, "commit")
		if idx > 0 {
			var buffer bytes.Buffer
			buffer.WriteString(baseUrl[0:idx])
			buffer.WriteString("pull/")
			buffer.WriteString(match[1])
			result := buffer.String()
			ci.PullURLRaw = &result

			log.WithFields(log.Fields{
				"baseURL":   baseUrl,
				"commitIdx": idx,
				"pullURL":   ci.PullURL(),
				"isPull":    ci.IsPull(),
			}).Debug("applyPullPropertiesChangeItem")
		}
	}
}

func shouldExcludeViaPullAttributes(pullId int, contextual *Contextual, parent *context.Context, c *model.Config) (*github.PullRequest, bool) {
	client := contextual.GetClient()
	timeout, cancel := contextual.CreateContext(parent)
	defer cancel()

	log.Debugf("Checking pull request %d", pullId)
	pr, _, e := client.PullRequests.Get(timeout, (*c).Owner, (*c).Repo, pullId)
	if e != nil || pr == nil {
		return nil, false
	}
	if c.ShouldExcludeByText(pr.Title) {
		return pr, true
	}
	for _, label := range pr.Labels {
		if c.ShouldExcludeByText(label.Name) {
			return pr, true
		}
	}
	return pr, false
}
