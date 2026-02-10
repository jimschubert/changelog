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
	"errors"
	"fmt"
	"strings"
	"time"
)

// ChangeItem stores properties exposed to users for Changelog creation
type ChangeItem struct {
	// The author of a commit
	AuthorRaw *string `json:"author"`

	// URL to author's GitHub profile
	AuthorURLRaw *string `json:"author_url"`

	// The commit title
	CommitMessageRaw *string `json:"commit_message"`

	// The commit date of the contribution (i.e. merge date)
	DateRaw *time.Time `json:"date"`

	// IsPullRaw determines if the commit was sourced from a pull request or directly committed to the branch
	IsPullRaw *bool `json:"is_pull"`

	// When IsPullRaw=true, this will point to the source of the pull request
	PullURLRaw *string `json:"pull_url"`

	// The commit's full SHA1 hash
	CommitHashRaw *string `json:"commit"`

	// The URL to the commit
	CommitURLRaw *string `json:"commit_url"`

	// An optional group identifier
	GroupRaw *string `json:"group"`
}

// Author or empty string
func (ci *ChangeItem) Author() string {
	if ci.AuthorRaw != nil {
		return *ci.AuthorRaw
	}

	return ""
}

// AuthorURL or empty string
func (ci *ChangeItem) AuthorURL() string {
	if ci.AuthorURLRaw != nil {
		return *ci.AuthorURLRaw
	}

	return ""
}

// Title is the first line of a commit message, otherwise empty string
func (ci *ChangeItem) Title() string {
	if ci.CommitMessageRaw != nil {
		idx := strings.Index(*ci.CommitMessageRaw, "\n")
		if idx > 0 {
			return (*ci.CommitMessageRaw)[0:idx]
		}

		return *ci.CommitMessageRaw
	}

	return ""
}

// Date or now
func (ci *ChangeItem) Date() time.Time {
	if ci.DateRaw != nil {
		return *ci.DateRaw
	}

	return time.Time{}
}

// IsPull or false
func (ci *ChangeItem) IsPull() bool {
	if ci.IsPullRaw != nil {
		return *ci.IsPullRaw
	}
	return false
}

// PullURL or empty string
func (ci *ChangeItem) PullURL() string {
	if ci.PullURLRaw != nil {
		return *ci.PullURLRaw
	}
	return ""
}

// CommitHash or empty string
func (ci *ChangeItem) CommitHash() string {
	if ci.CommitHashRaw != nil {
		return *ci.CommitHashRaw
	}
	return ""
}

// CommitURL or empty string
func (ci *ChangeItem) CommitURL() string {
	if ci.CommitURLRaw != nil {
		return *ci.CommitURLRaw
	}
	return ""
}

// CommitHashShort is first 10 characters of CommitHash, or CommitHash if it's already short
func (ci *ChangeItem) CommitHashShort() string {
	hash := ci.CommitHash()
	if len(hash) > 10 {
		return hash[0:10]
	}
	return hash
}

// PullID is the numerical ID of a pull request, extracted from PullURL
func (ci *ChangeItem) PullID() (string, error) {
	url := ci.PullURL()
	if len(url) > 0 {
		parts := strings.Split(url, "/")
		id := parts[len(parts)-1]
		return id, nil
	}

	return "", errors.New("no pull url available")
}

// Group is the targeted group for a commit, or empty string
func (ci *ChangeItem) Group() string {
	if ci.GroupRaw != nil {
		return *ci.GroupRaw
	}
	return ""
}

// GoString displays debuggable format of ChangeItem
func (ci *ChangeItem) GoString() string {
	var builder strings.Builder

	builder.WriteString("ChangeItem: {")
	builder.WriteString(" Commit: ")
	builder.WriteString(ci.CommitHashShort())
	builder.WriteString(", Author: ")
	builder.WriteString(ci.Author())
	builder.WriteString(", Time: ")
	builder.WriteString(fmt.Sprintf("%v", ci.Date()))
	builder.WriteString(", CommitMessage: ")
	builder.WriteString(ci.Title())
	builder.WriteString(" }")

	return builder.String()
}
