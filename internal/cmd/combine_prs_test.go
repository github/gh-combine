package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/github/gh-combine/internal/github"
	"github.com/stretchr/testify/assert"
)

func TestCreatePullRequest(t *testing.T) {
	client := &MockRESTClient{
		PostFunc: func(endpoint string, body interface{}, response interface{}) error {
			if strings.Contains(endpoint, "/pulls") {
				if prResponse, ok := response.(*struct{ Number int }); ok {
					prResponse.Number = 123 // Mock PR number
				}
			} else if strings.Contains(endpoint, "/labels") || strings.Contains(endpoint, "/assignees") {
				// Mock successful label/assignee addition
				return nil
			}
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
