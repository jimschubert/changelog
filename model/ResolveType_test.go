package model

import "testing"

func TestResolveType_String(t *testing.T) {
	tests := []struct {
		name string
		r    ResolveType
		want string
	}{
		{ "Commits.String()", Commits, "commits" },
		{ "PullRequests.String()", PullRequests, "pulls" },
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
		{ "unmarshal commits", Commits, args{b:[]byte("\"commits\"")}, false },
		{ "unmarshal prs", PullRequests, args{b:[]byte("\"prs\"")}, false },
		{ "unmarshal pulls", PullRequests, args{b:[]byte("\"pulls\"")}, false },
		{ "unmarshal pullrequest", PullRequests, args{b:[]byte("\"pullrequest\"")}, false },
		{ "unmarshal commits (bad single character)", Commits, args{b:[]byte("\"a\"")}, true },
		{ "unmarshal prs (bad single character)", PullRequests, args{b:[]byte("\"1\"")}, true },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}