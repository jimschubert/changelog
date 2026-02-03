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

func TestSortDirection_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		s       SortDirection
		want    []byte
		wantErr bool
	}{
		{"marshal ascending", Ascending, []byte(`"asc"`), false},
		{"marshal descending", Descending, []byte(`"desc"`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalJSON()
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

func TestSortDirection_Ptr(t *testing.T) {
	d := Descending
	a := Ascending
	tests := []struct {
		name string
		s    SortDirection
		want *SortDirection
	}{
		{"pointer to descending", d, &d},
		{"pointer to ascending", a, &a},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Ptr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ptr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortDirection_String(t *testing.T) {
	tests := []struct {
		name string
		s    SortDirection
		want string
	}{
		{"Ascending.String()", Ascending, "asc"},
		{"Descending.String()", Descending, "desc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortDirection_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		s       SortDirection
		args    args
		wantErr bool
	}{
		{"unmarshal asc", Ascending, args{b: []byte(`"asc"`)}, false},
		{"unmarshal ASC", Ascending, args{b: []byte(`"ASC"`)}, false},
		{"unmarshal ascending", Ascending, args{b: []byte(`"ascending"`)}, false},

		{"unmarshal desc", Descending, args{b: []byte(`"desc"`)}, false},
		{"unmarshal DESC", Descending, args{b: []byte(`"DESC"`)}, false},
		{"unmarshal descending", Descending, args{b: []byte(`"descending"`)}, false},

		{"unmarshal unknown to descending", Descending, args{b: []byte(`"unk"`)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSortDirection_UnmarshalYAML(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		s       SortDirection
		args    args
		wantErr bool
	}{
		{"unmarshal asc", Ascending, args{b: []byte(`asc`)}, false},
		{"unmarshal ASC", Ascending, args{b: []byte(`ASC`)}, false},
		{"unmarshal ascending", Ascending, args{b: []byte(`ascending`)}, false},

		{"unmarshal desc", Descending, args{b: []byte(`desc`)}, false},
		{"unmarshal DESC", Descending, args{b: []byte(`DESC`)}, false},
		{"unmarshal descending", Descending, args{b: []byte(`descending`)}, false},

		{"unmarshal unknown to descending", Descending, args{b: []byte(`unk`)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UnmarshalYAML(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
