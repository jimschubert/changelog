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
	"reflect"
	"testing"
	"time"
)

func p(s string) *string {
	return &s
}

func pTime(t time.Time) *time.Time {
	return &t
}

func TestChangeItem_getAuthor(t *testing.T) {
	type fields struct {
		Author    *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty Author", fields{Author: nil}, ""},
		{"populated Author", fields{Author: p("jimschubert")}, "jimschubert"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				AuthorRaw: tt.fields.Author,
			}
			if got := ci.Author(); got != tt.want {
				t.Errorf("Author() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getAuthorURL(t *testing.T) {
	type fields struct {
		AuthorURL *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty AuthorURL", fields{AuthorURL: nil}, ""},
		{"populated AuthorURL", fields{AuthorURL: p("https://github.com/jimschubert")}, "https://github.com/jimschubert"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				AuthorURLRaw: tt.fields.AuthorURL,
			}
			if got := ci.AuthorURL(); got != tt.want {
				t.Errorf("AuthorURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getCommit(t *testing.T) {
	type fields struct {
		Commit    *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty CommitHash", fields{Commit: nil}, ""},
		{"populated CommitHash", fields{Commit: p("4b825dc642cb6eb9a060e54bf8d69288fbee4904")}, "4b825dc642cb6eb9a060e54bf8d69288fbee4904"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				CommitHashRaw: tt.fields.Commit,
			}
			if got := ci.CommitHash(); got != tt.want {
				t.Errorf("CommitHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getDate(t *testing.T) {
	d := time.Date(2020, time.February, 20, 10, 10, 2, 20, &time.Location{})
	type fields struct {
		Date      *time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Time
	}{
		{"empty Date", fields{Date: nil}, time.Time{}},
		{"populated Date", fields{Date: pTime(d)}, d},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				DateRaw: tt.fields.Date,
			}
			if got := ci.Date(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Date() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getIsPull(t *testing.T) {
	f := false
	tt := true
	type fields struct {
		IsPull    *bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"empty IsPull", fields{IsPull: nil}, false},
		{"populated IsPull (false)", fields{IsPull: &f}, false},
		{"populated IsPull (true)", fields{IsPull: &tt}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				IsPullRaw: tt.fields.IsPull,
			}
			if got := ci.IsPull(); got != tt.want {
				t.Errorf("IsPull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getPullID(t *testing.T) {
	pull := "https://github.com/OpenAPITools/openapi-generator/pull/5472"
	empty := ""
	type fields struct {
		PullURL   *string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"nil PullID", fields{PullURL: nil}, "", true},
		{"populated PullID", fields{PullURL: &pull}, "5472", false},
		{"empty PullID", fields{PullURL: &empty}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				PullURLRaw: tt.fields.PullURL,
			}
			got, err := ci.PullID()
			if (err != nil) != tt.wantErr {
				t.Errorf("PullID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PullID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getPullURL(t *testing.T) {
	pull := "https://github.com/OpenAPITools/openapi-generator/pull/5472"
	type fields struct {
		PullURL   *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty PullID", fields{PullURL: nil}, ""},
		{"populated PullID", fields{PullURL: &pull}, pull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				PullURLRaw: tt.fields.PullURL,
			}
			if got := ci.PullURL(); got != tt.want {
				t.Errorf("PullURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getGroup(t *testing.T) {
	type fields struct {
		Group     *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty Group", fields{Group: nil}, ""},
		{"populated Group", fields{Group: p("docs")}, "docs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				GroupRaw: tt.fields.Group,
			}
			if got := ci.Group(); got != tt.want {
				t.Errorf("Group() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_getCommitHashShort(t *testing.T) {
	type fields struct {
		CommitHash *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty CommitHash", fields{CommitHash: nil}, ""},
		{"populated CommitHash", fields{CommitHash: p("4b825dc642cb6eb9a060e54bf8d69288fbee4904")}, "4b825dc642"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				CommitHashRaw: tt.fields.CommitHash,
			}
			if got := ci.CommitHashShort(); got != tt.want {
				t.Errorf("CommitHashShort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_CommitURL(t *testing.T) {
	type fields struct {
		CommitURLRaw     *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{ "empty CommitURL", fields{CommitURLRaw: nil}, ""},
		{ "populated CommitURL", fields{CommitURLRaw: p("https://github.com/jimschubert/changelog/commit/6db2267d8fbc2929655884825a76856bc3244acd")}, "https://github.com/jimschubert/changelog/commit/6db2267d8fbc2929655884825a76856bc3244acd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				CommitURLRaw: tt.fields.CommitURLRaw,
			}
			if got := ci.CommitURL(); got != tt.want {
				t.Errorf("CommitURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_Title(t *testing.T) {
	type fields struct {
		CommitMessageRaw *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty Title from nil CommitMessage", fields{CommitMessageRaw: nil}, ""},
		{"empty Title from empty CommitMessage", fields{CommitMessageRaw: p("")}, ""},
		{"empty Title from single-line CommitMessage", fields{CommitMessageRaw: p("Initial Commit")}, "Initial Commit"},
		{"empty Title from multi-line CommitMessage",
			fields{CommitMessageRaw: p(
				`Initial Commit
					
					This is the initial commit
					`)},
			"Initial Commit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				CommitMessageRaw: tt.fields.CommitMessageRaw,
			}
			if got := ci.Title(); got != tt.want {
				t.Errorf("Title() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeItem_GoString(t *testing.T) {
	type fields struct {
		AuthorRaw        *string
		AuthorURLRaw     *string
		CommitMessageRaw *string
		DateRaw          *time.Time
		IsPullRaw        *bool
		PullURLRaw       *string
		CommitHashRaw    *string
		CommitURLRaw     *string
		GroupRaw         *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"GoString should not error when fields are nil", fields{}, "ChangeItem: { Commit: , Author: , Time: 0001-01-01 00:00:00 +0000 UTC, CommitMessage:  }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &ChangeItem{
				AuthorRaw:        tt.fields.AuthorRaw,
				AuthorURLRaw:     tt.fields.AuthorURLRaw,
				CommitMessageRaw: tt.fields.CommitMessageRaw,
				DateRaw:          tt.fields.DateRaw,
				IsPullRaw:        tt.fields.IsPullRaw,
				PullURLRaw:       tt.fields.PullURLRaw,
				CommitHashRaw:    tt.fields.CommitHashRaw,
				CommitURLRaw:     tt.fields.CommitURLRaw,
				GroupRaw:         tt.fields.GroupRaw,
			}
			if got := ci.GoString(); got != tt.want {
				t.Errorf("GoString() = %v, want %v", got, tt.want)
			}
		})
	}
}