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

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
	ResolveType *ResolveType `json:"resolve"`

	// The Owner (user or org) of the target repository
	Owner string `json:"owner"`

	// The target repository
	Repo string `json:"repo"`

	// A set of Grouping objects which allow to define groupings for changelog output.
	// Commits are associated with the first matching group.
	Groupings *[]Grouping `json:"groupings"`

	// As set of square-bracket regex patterns, wrapped texts and/or labels to be excluded from output.
	// If the commit message or pr labels reference any text in this Exclude set, that commit
	// will be ignored./**/
	Exclude *[]string `json:"exclude"`

	// Optional base url when targeting GitHub Enterprise
	Enterprise *string `json:"enterprise"`

	// Custom template following Go text/template syntax
	// For more details, see https://golang.org/pkg/text/template/
	Template *string `json:"template"`

	// SortDirection defines the order of commits within the changelog
	SortDirection *SortDirection `json:"sort"`
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

	return json.Unmarshal(b, c)
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
	return fmt.Sprintf(buffer.String())
}

// LoadOrNewConfig will attempt to load path, otherwise returns a newly constructed config.
func LoadOrNewConfig(path *string, owner string, repo string) *Config {
	defaultResolveType := Commits
	defaultSortDirection := Descending
	config := Config{
		SortDirection: &defaultSortDirection,
		ResolveType: &defaultResolveType,
	}
	if path != nil {
		err := config.Load(*path)
		if err == nil {
			if config.Owner == "" || strings.Compare(owner, config.Owner) != 0 {
				config.Owner = owner
			}
			if config.Repo == "" || strings.Compare(repo, config.Repo) != 0 {
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
