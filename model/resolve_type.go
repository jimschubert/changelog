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
	"fmt"
	"strings"
)

// ResolveType is a type alias representing the enumeration of options
// which configure how commits are processed (if commit only or if we lookup any available pull request info)
type ResolveType uint8

const (
	// Commits only
	Commits ResolveType = 1 << iota
	// PullRequests requests that we pull PR information if available
	PullRequests ResolveType = 1 << iota
)

// MarshalJSON converts ResolveType into a string representation sufficient for JSON
func (r *ResolveType) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte(""), nil
	}

	it := *r
	switch it {
	case PullRequests:
		return []byte(`"pulls"`), nil
	case Commits:
		fallthrough
	default:
		return []byte(`"commits"`), nil
	}
}

// UnmarshalJSON converts a JSON formatted character array into ResolveType
func (r *ResolveType) UnmarshalJSON(b []byte) error {
	if len(b) == 1 {
		*r = ResolveType(b[0])
		return nil
	}

	s := strings.TrimSpace(string(b))
	switch s {
	case "commits", `"commits"`:
		*r = Commits
	case "pulls", "pullrequest", "prs", `"pulls"`, `"pullrequest"`, `"prs"`:
		*r = PullRequests
	default:
		return fmt.Errorf("unknown resolve type %q", s)
	}
	return nil
}

func (r *ResolveType) UnmarshalYAML(b []byte) error {
	return r.UnmarshalJSON(b)
}

func (r *ResolveType) MarshalYAML() ([]byte, error) {
	return r.MarshalJSON()
}

// String displays a human readable representation of the ResolveType values
func (r ResolveType) String() string {
	switch r {
	case PullRequests:
		return "pulls"
	case Commits:
		fallthrough
	default:
		return "commits"
	}
}

func (r ResolveType) Ptr() *ResolveType {
	return &r
}
