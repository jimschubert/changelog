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
	"fmt"
	"strings"
	"time"
)

// ChangeItem stores properties exposed to users for Changelog creation
type ChangeItem struct {
	// The author of a commit
	Author_ *string `json:"author"`

	// URL to author's GitHub profile
	AuthorURL_ *string `json:"author_url"`

	// The commit title
	CommitMessage_ *string `json:"commit_message"`

	// The commit date of the contribution (i.e. merge date)
	Date_ *time.Time `json:"date"`

	// IsPull_ determines if the commit was sourced from a pull request or directly committed to the branch
	IsPull_ *bool `json:"is_pull"`

	// When IsPull_=true, this will point to the source of the pull request
	PullURL_ *string `json:"pull_url"`

	// The commit's full SHA1 hash
	CommitHash_ *string `json:"commit"`

	// The URL to the commit
	CommitURL_ *string `json:"commit_url"`

	// An optional group identifier
	Group_ *string `json:"group"`
}

func (ci *ChangeItem) Author() string {
	if ci.Author_ != nil {
		return *ci.Author_
	}

	return ""
}

func (ci *ChangeItem) AuthorURL() string {
	if ci.AuthorURL_ != nil {
		return *ci.AuthorURL_
	}

	return ""
}

func (ci *ChangeItem) Title() string {
	if ci.CommitMessage_ != nil {
		idx := strings.Index(*ci.CommitMessage_, "\n")
		if idx > 0 {
			return (*ci.CommitMessage_)[0:idx]
		}

		return *ci.CommitMessage_
	}

	return ""
}

func (ci *ChangeItem) Date() time.Time {
	if ci.Date_ != nil {
		return *ci.Date_
	}

	return time.Time{}
}

func (ci *ChangeItem) IsPull() bool {
	if ci.IsPull_ != nil {
		return *ci.IsPull_
	}
	return false
}

func (ci *ChangeItem) PullURL() string {
	if ci.PullURL_ != nil {
		return *ci.PullURL_
	}
	return ""
}

func (ci *ChangeItem) CommitHash() string {
	if ci.CommitHash_ != nil {
		return *ci.CommitHash_
	}
	return ""
}

func (ci *ChangeItem) CommitURL() string {
	if ci.CommitURL_ != nil {
		return *ci.CommitURL_
	}
	return ""
}

func (ci *ChangeItem) CommitHashShort() string {
	hash := ci.CommitHash()
	if len(hash) > 10 {
		return hash[0:10]
	}
	return hash
}

func (ci *ChangeItem) PullID() (string, error) {
	url := ci.PullURL()
	if len(url) > 0 {
		parts := strings.Split(url, "/")
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("no pull url available")
}

func (ci *ChangeItem) Group() string {
	if ci.Group_ != nil {
		return *ci.Group_
	}
	return ""
}

func (ci *ChangeItem) GoString() string {
	var buffer bytes.Buffer

	buffer.WriteString("ChangeItem: {")
	buffer.WriteString(" Commit: ")
	buffer.WriteString(ci.CommitHashShort())
	buffer.WriteString(", Author_: ")
	buffer.WriteString(ci.Author())
	buffer.WriteString(", Time: ")
	buffer.WriteString(fmt.Sprintf("%v", ci.Date()))
	buffer.WriteString(", CommitMessage_: ")
	buffer.WriteString(ci.Title())
	buffer.WriteString(" }")

	return buffer.String()
}
