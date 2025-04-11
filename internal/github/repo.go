package github

import (
	"errors"
	"fmt"
	"strings"
)

type Repo struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

var ErrInvalidRepository = errors.New("invalid repository")

func ParseRepo(s string) (Repo, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Repo{}, fmt.Errorf("%w: %s", ErrInvalidRepository, s)
	}

	return Repo{
		Owner: parts[0],
		Repo:  parts[1],
	}, nil
}

func (r Repo) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Repo)
}

func (r Repo) PullsEndpoint() string {
	return fmt.Sprintf("repos/%s/%s/pulls?state=open", r.Owner, r.Repo)
}
