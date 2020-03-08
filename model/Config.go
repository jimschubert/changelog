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
)

// Config provides a user with more robust options for Changelog configuration
type Config struct {
	// Defines whether we resolve commits only or query additional information from pull requests
	*ResolveType `json:"resolve"`

	// The Owner (user or org) of the target repository
	Owner string `json:"owner"`

	// The target repository
	Repo string `json:"repo"`

	// A set of square-bracket wrapped texts and/or labels (if resolving pull requests) we will define groupings for
	// Commits are associated with the first matching group.
	Groupings *[]string `json:"groupings"`

	// As set of square-bracket wrapped texts and/or labels to be excluded from output.
	// If the commit message or pr labels reference any text in this Exclude set, that commit
	// will be ignored./**/
	Exclude *[]string `json:"exclude"`

	// Optional base url when targeting GitHub Enterprise
	Enterprise *string `json:"enterprise"`

	// Custom template following Go text/template syntax
	// For more details, see https://golang.org/pkg/text/template/
	Template *string `json:"template"`
}

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

func (c *Config) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Config:")
	buffer.WriteString(fmt.Sprintf("\n\tResolveType: %s", c.ResolveType))
	buffer.WriteString("\n\tOwner: ")
	buffer.WriteString(c.Owner)
	buffer.WriteString("\n\tRepo: ")
	buffer.WriteString(c.Repo)
	buffer.WriteString(fmt.Sprintf("\n\tGroupings: %v", c.Groupings))
	buffer.WriteString(fmt.Sprintf("\n\tExclude: %v", c.Exclude))
	buffer.WriteString(fmt.Sprintf("\n\tEnterprise: %v", c.Enterprise))
	buffer.WriteString(fmt.Sprintf("\n\tTemplate: %v", c.Template))

	return fmt.Sprintf(buffer.String())
}

func LoadOrNewConfig(path *string, owner string, repo string) *Config {
	defaultResolveType := Commits
	config := Config{}
	if path != nil {
		err := config.Load(*path)
		if err == nil {
			return &config
		}
	}

	config.Owner = owner
	config.Repo = repo
	config.ResolveType = &defaultResolveType

	return &config
}