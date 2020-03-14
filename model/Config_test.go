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
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func createTempConfig(t *testing.T, data string) (fileLocation string, cleanup func()) {
	t.Helper()
	tempDir, err := ioutil.TempDir("", "Config_test")
	if err != nil {
		t.Fatal(err)
	}
	r := rand.Int()
	testHash := hash(t.Name())
	testFile := fmt.Sprintf("file-%d-%d.json", r, testHash)
	filePath := filepath.Join(tempDir, testFile)
	if err := ioutil.WriteFile(filePath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	return filePath, func() { _ = os.RemoveAll(filePath) }
}

func ptrResolveType(t ResolveType) *ResolveType {
	return &t
}

func ptrStringArray(s ...string) *[]string {
	arr := make([]string, 0)
	if len(s) > 0 {
		arr = append(arr, s...)
	}
	return &arr
}

func TestConfig_Load(t *testing.T) {
	type fields struct {
		JSONData    string
		ResolveType *ResolveType
		Owner       string
		Repo        string
		Groupings   *[]string
		Exclude     *[]string
		Enterprise  *string
		Template    *string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"Loads valid empty json", fields{JSONData: "{}"}, false},
		{"Loads valid json resolve-only", fields{JSONData: `{"resolve": "commits"}`, ResolveType: ptrResolveType(Commits)}, false},
		{"Fail on valid json with invalid data type resolve-only", fields{JSONData: `{"resolve": 1.0}`}, true}, // note that 1 would resolve since enum is an int
		{"Loads valid json owner-only", fields{JSONData: `{"owner": "jimschubert"}`, Owner: "jimschubert"}, false},
		{"Fail on valid json with invalid data type owner-only", fields{JSONData: `{"owner": []}`}, true},
		{"Loads valid json repo-only", fields{JSONData: `{"repo": "changelog"}`, Owner: "changelog"}, false},
		{"Fail on valid json with invalid data type repo-only", fields{JSONData: `{"repo": []}`}, true},
		{"Loads valid json groupings-only", fields{JSONData: `{"groupings": []}`, Groupings: ptrStringArray()}, false},
		{"Fail on valid json with invalid data type groupings-only", fields{JSONData: `{"groupings": 4}`}, true},
		{"Loads valid json exclude-only", fields{JSONData: `{"exclude": []}`, Exclude: ptrStringArray()}, false},
		{"Fail on valid json with invalid data type exclude-only", fields{JSONData: `{"exclude": 1}`}, true},
		{"Loads valid json enterprise-only", fields{JSONData: `{"enterprise": "https://ghe.example.com"}`, Enterprise: p("https://ghe.example.com")}, false},
		{"Fail on valid json with invalid data type enterprise-only", fields{JSONData: `{"enterprise": 0}`}, true},
		{"Loads valid json template-only", fields{JSONData: `{"template": "/path/to/template"}`, Template: p("/path/to/template")}, false},
		{"Fail on valid json with invalid data type template-only", fields{JSONData: `{"template": []}`}, true},
		{"Loads valid full json",
			fields{
				JSONData:    `{"resolve":"commits","owner":"jimschubert","repo":"ossify","groupings":["feature","bug"],"exclude":["wip","help wanted"],"enterprise":"https://ghe.example.com","template":"/path/to/template"}`,
				ResolveType: ptrResolveType(Commits),
				Owner:       "jimschubert",
				Repo:        "ossify",
				Groupings:   ptrStringArray("feature", "bug"),
				Exclude:     ptrStringArray("wip", "help wanted"),
				Enterprise:  p("https://ghe.example.com"),
				Template:    p("/path/to/template"),
			}, false},
		{"Fails on invalid json",
			fields{
				JSONData: `{"resolve":"commits":"owner":"jimschubert"}`,
			}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonLocation, cleanup := createTempConfig(t, tt.fields.JSONData)
			defer cleanup()
			t.Run(tt.name, func(t *testing.T) {
				c := &Config{
					ResolveType: tt.fields.ResolveType,
					Owner:       tt.fields.Owner,
					Repo:        tt.fields.Repo,
					Groupings:   tt.fields.Groupings,
					Exclude:     tt.fields.Exclude,
					Enterprise:  tt.fields.Enterprise,
					Template:    tt.fields.Template,
				}
				if err := c.Load(jsonLocation); (err != nil) != tt.wantErr {
					t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestConfig_String(t *testing.T) {
	type fields struct {
		ResolveType *ResolveType
		Owner       string
		Repo        string
		Groupings   *[]string
		Exclude     *[]string
		Enterprise  *string
		Template    *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"outputs string",
			fields{
				ResolveType: ptrResolveType(Commits),
				Owner:       "jimschubert",
				Repo:        "ossify",
				Groupings:   ptrStringArray("feature", "bug"),
				Exclude:     ptrStringArray("wip", "help wanted"),
				Enterprise:  p("https://ghe.example.com"),
				Template:    p("/path/to/template"),
			}, `Config: { ResolveType: commits Owner: jimschubert Repo: ossify Groupings: &[feature bug] Exclude: &[wip help wanted] Enterprise: https://ghe.example.com Template: /path/to/template Sort:  }`},

		{"outputs string for nil properties",
			fields{
			}, `Config: { ResolveType: <nil> Owner:  Repo:  Groupings: <nil> Exclude: <nil> Enterprise:  Template:  Sort:  }`},
	}
	// Config: {ResolveType: commits Owner: jimschubert Repo: ossify Groupings: &[feature bug] Exclude: &[wip help wanted] Enterprise: 0xc00003c5b0 Template: 0xc00003c5c0}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ResolveType: tt.fields.ResolveType,
				Owner:       tt.fields.Owner,
				Repo:        tt.fields.Repo,
				Groupings:   tt.fields.Groupings,
				Exclude:     tt.fields.Exclude,
				Enterprise:  tt.fields.Enterprise,
				Template:    tt.fields.Template,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadOrNewConfig(t *testing.T) {
	type args struct {
		path  *string
		owner string
		repo  string
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadOrNewConfig(tt.args.path, tt.args.owner, tt.args.repo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadOrNewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
