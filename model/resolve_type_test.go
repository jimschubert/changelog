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
	"reflect"
	"testing"
)

func TestResolveType_String(t *testing.T) {
	tests := []struct {
		name string
		r    ResolveType
		want string
	}{
		{"Commits.String()", Commits, "commits"},
		{"PullRequests.String()", PullRequests, "pulls"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveType_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		r       ResolveType
		args    args
		wantErr bool
	}{
		{"unmarshal commits", Commits, args{b: []byte("\"commits\"")}, false},
		{"unmarshal prs", PullRequests, args{b: []byte("\"prs\"")}, false},
		{"unmarshal pulls", PullRequests, args{b: []byte("\"pulls\"")}, false},
		{"unmarshal pullrequest", PullRequests, args{b: []byte("\"pullrequest\"")}, false},
		{"unmarshal commits (bad single character)", Commits, args{b: []byte("\"a\"")}, true},
		{"unmarshal prs (bad single character)", PullRequests, args{b: []byte("\"1\"")}, true},
		{"unmarshal prs (bad empty character)", PullRequests, args{b: []byte("")}, true},
		{"unmarshal prs (nil for null terminating character)", PullRequests, args{b: []byte("\x00")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		r       ResolveType
		want    []byte
		wantErr bool
	}{
		{"marshal commits", Commits, []byte("\"commits\""), false},
		{"marshal pulls", PullRequests, []byte("\"pulls\""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveType_MarshalJSON1(t *testing.T) {
	tests := []struct {
		name    string
		r       *ResolveType
		want    []byte
		wantErr bool
	}{
		{"marshal commits on nil instance succeeds with empty result", nil, []byte(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveType_UnmarshalYAML(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		r       ResolveType
		args    args
		wantErr bool
	}{
		{"unmarshal commits", Commits, args{b: []byte("commits")}, false},
		{"unmarshal prs", PullRequests, args{b: []byte("prs")}, false},
		{"unmarshal pulls", PullRequests, args{b: []byte("pulls")}, false},
		{"unmarshal pullrequest", PullRequests, args{b: []byte("pullrequest")}, false},
		{"unmarshal commits (bad single character)", Commits, args{b: []byte("\"a\"")}, true},
		{"unmarshal prs (bad single character)", PullRequests, args{b: []byte("\"1\"")}, true},
		{"unmarshal prs (bad empty character)", PullRequests, args{b: []byte("")}, true},
		{"unmarshal prs (nil for null terminating character)", PullRequests, args{b: []byte("\x00")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.UnmarshalYAML(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
