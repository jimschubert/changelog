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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
)

// Grouping allows assigning a grouping name with a set of regex patterns or texts.
// These patterns are evaluated against commit titles and, if resolving pull requests, labels.
type Grouping struct {
	// Name of the group, displayed in changelog output
	Name string `json:"name"`

	// Patterns to be evaluated for association in this group
	Patterns []string `json:"patterns"`
}

// Config provides a user with more robust options for Changelog configuration
type Config struct {
	// Defines whether we resolve commits only or query additional information from pull requests
	ResolveType *ResolveType `json:"resolve" yaml:"resolve"`

	// The Owner (user or org) of the target repository
	Owner string `json:"owner"`

	// The target repository
	Repo string `json:"repo"`

	// A set of Grouping objects which allow to define groupings for changelog output.
	// Commits are associated with the first matching group.
	Groupings *[]Grouping `json:"groupings,omitempty"`

	// As set of square-bracket regex patterns, wrapped texts and/or labels to be excluded from output.
	// If the commit message or pr labels reference any text in this Exclude set, that commit
	// will be ignored./**/
	Exclude *[]string `json:"exclude,omitempty"`

	// Optional base url when targeting GitHub Enterprise
	Enterprise *string `json:"enterprise,omitempty"`

	// Custom template following Go text/template syntax
	// For more details, see https://golang.org/pkg/text/template/
	Template *string `json:"template,omitempty"`

	// SortDirection defines the order of commits within the changelog
	SortDirection *SortDirection `json:"sort"`

	// PreferLocal defines whether commits may be queried locally. Requires executing from within a Git repository.
	PreferLocal *bool `json:"local,omitempty"`

	// MaxCommits defines the maximum number of commits to be processed.
	MaxCommits *int `json:"max_commits,omitempty"`
}

// Load a Config from path
func (c *Config) Load(path string) error {
	b, err := ioutil.ReadFile(path)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if strings.HasSuffix(path, ".json") {
		return json.Unmarshal(b, c)
	}

	return yaml.Unmarshal(b, c)
}

// GetPreferLocal returns the user-specified preference for local commit querying, otherwise the default of 'false'
func (c *Config) GetPreferLocal() bool {
	if c.PreferLocal == nil {
		return false
	}

	return *c.PreferLocal
}

// GetMaxCommits returns the user-specified preference for maximum commit count, otherwise the default of 500
func (c *Config) GetMaxCommits() int {
	if c.MaxCommits == nil {
		// This default matches the maximum defined in GitHub's compare API.
		// see https://developer.github.com/v3/repos/commits/#compare-two-commits
		return 250
	}

	return *c.MaxCommits
}

func (c *Config) ShouldExcludeByText(text *string) bool {
	if text == nil || c.Exclude == nil || len(*c.Exclude) == 0 {
		return false
	}
	for _, pattern := range *c.Exclude {
		re := regexp.MustCompile(pattern)
		if re.Match([]byte(*text)) {
			log.WithFields(log.Fields{"text": *text, "pattern": pattern}).Debug("exclude via pattern")
			return true
		}
	}
	return false
}

func (c *Config) FindGroup(commitMessage string) *string {
	var grouping *string
	if c.Groupings != nil && len(*c.Groupings) > 0 {
		title := strings.Split(commitMessage, "\n")[0]
		for _, g := range *c.Groupings {
			for _, pattern := range g.Patterns {
				re := regexp.MustCompile(pattern)
				if re.Match([]byte(title)) {
					grouping = &g.Name
					log.WithFields(log.Fields{"grouping": *grouping, "title": title}).Debug("found group name for commit")
					return grouping
				}
			}
		}
	}
	return grouping
}

// String displays a human readable representation of a Config
func (c *Config) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Config: { ")
	buffer.WriteString(fmt.Sprintf(`ResolveType: %s`, c.ResolveType))
	buffer.WriteString(fmt.Sprintf(` Owner: %s`, c.Owner))
	buffer.WriteString(fmt.Sprintf(` Repo: %s`, c.Repo))
	buffer.WriteString(fmt.Sprintf(` Groupings: %v`, c.Groupings))
	buffer.WriteString(fmt.Sprintf(` Exclude: %v`, c.Exclude))
	buffer.WriteString(" Enterprise: ")
	if c.Enterprise != nil {
		buffer.WriteString(*c.Enterprise)
	}
	buffer.WriteString(" Template: ")
	if c.Template != nil {
		buffer.WriteString(*c.Template)
	}
	buffer.WriteString(" Sort: ")
	if c.SortDirection != nil {
		buffer.WriteString((*c.SortDirection).String())
	}
	buffer.WriteString(" }")
	return buffer.String()
}

// LoadOrNewConfig will attempt to load path, otherwise returns a newly constructed config.
func LoadOrNewConfig(path *string, owner string, repo string) *Config {
	defaultResolveType := Commits
	defaultSortDirection := Descending
	config := Config{
		SortDirection: &defaultSortDirection,
		ResolveType:   &defaultResolveType,
	}
	if path != nil {
		err := config.Load(*path)
		if err == nil {
			if config.Owner == "" || (strings.Compare(owner, config.Owner) != 0 && owner != "") {
				config.Owner = owner
			}
			if config.Repo == "" || (strings.Compare(repo, config.Repo) != 0 && repo != "") {
				config.Repo = repo
			}
			if config.ResolveType == nil {
				config.ResolveType = &defaultResolveType
			}
			return &config
		}
	}

	config.Owner = owner
	config.Repo = repo
	config.ResolveType = &defaultResolveType

	return &config
}
