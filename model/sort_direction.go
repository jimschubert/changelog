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

type SortDirection uint8

const (
	// Descending means most recent commits are at the top of the changelog
	Descending SortDirection = 1 << iota

	// Ascending means earlier commits are at the top of the changelog, more recent are at the bottom
	Ascending SortDirection = 1 << iota
)

// MarshalJSON converts SortDirection into a string representation sufficient for JSON
func (s *SortDirection) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte(""), nil
	}
	it := *s
	switch it {
	case Ascending:
		return []byte(`"asc"`), nil
	case Descending:
		fallthrough
	default:
		return []byte(`"desc"`), nil
	}
}

// UnmarshalJSON converts a JSON formatted character array into SortDirection
func (s *SortDirection) UnmarshalJSON(b []byte) error {
	if len(b) == 1 {
		*s = SortDirection(b[0])
		return nil
	}
	it := string(b)
	switch it {
	case `asc`, `ascending`, `ASC`, `"asc"`, `"ascending"`, `"ASC"`:
		*s = Ascending
	case `desc`, `descending`, `DESC`, `"desc"`, `"descending"`, `"DESC"`:
		fallthrough
	default:
		*s = Descending
	}
	return nil
}

func (s *SortDirection) UnmarshalYAML(b []byte) error {
	return s.UnmarshalJSON(b)
}

func (s *SortDirection) MarshalYAML() ([]byte, error) {
	return s.MarshalJSON()
}

// String displays a human readable representation of the SortDirection values
func (s SortDirection) String() string {
	switch s {
	case Ascending:
		return "asc"
	case Descending:
		fallthrough
	default:
		return "desc"
	}
}

func (s SortDirection) Ptr() *SortDirection {
	return &s
}
