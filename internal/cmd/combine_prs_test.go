package cmd

import (
	"context"
	"testing"

	"github.com/github/gh-combine/internal/github"
	"github.com/stretchr/testify/assert"
)

func TestCreatePullRequest(t *testing.T) {
	client := &MockRESTClient{
		PostFunc: func(endpoint string, body interface{}, response interface{}) error {
			return nil
		},
	}
	repo := github.Repo{
		Owner: "test-owner",
		Repo:  "test-repo",
	}
	title := "Test PR"
	head := "test-branch"
	base := "main"
	body := "This is a test PR."
	labels := []string{"bug", "enhancement"}
	assignees := []string{"octocat", "hubot"}

	err := createPullRequest(context.Background(), client, repo, title, head, base, body, labels, assignees)
	assert.NoError(t, err)
}
