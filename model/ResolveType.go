package model

import "fmt"

type ResolveType uint8

const (
	Commits      ResolveType = 1 << iota
	PullRequests ResolveType = 1 << iota
)

func (r *ResolveType) UnmarshalJSON(b []byte) error {
	if len(b) == 1 {
		*r = ResolveType(b[0])
		return nil
	}

	s := string(b)
	switch s {
	case `"commits"`:
		*r = Commits
	case `"pulls"`, `"pullrequest"`, `"prs"`:
		*r = PullRequests
	default:
		return fmt.Errorf("unknown resolve type %q", s)
	}
	return nil
}

func (r ResolveType) String() string {
	switch r {
	case Commits:
		return "commits"
	case PullRequests:
		return "pulls"
	default:
		panic("code is unreachable")
	}
}
