package github

import (
	"fmt"
	"strings"
)

type Repo struct {
	Owner string
	Repo  string
}

// @grant: testing this should be pretty easy.
// Make sure to create a custom error.
func ParseRepo(repo string) (Repo, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return Repo{}, fmt.Errorf("invalid repository format: %s", repo)
	}

	return Repo{
		Owner: parts[0],
		Repo:  parts[1],
	}, nil
}
