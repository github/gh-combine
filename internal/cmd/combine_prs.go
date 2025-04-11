package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/github/gh-combine/internal/github"
)

func CombinePRs(ctx context.Context, graphQlClient *api.GraphQLClient, restClient *api.RESTClient, repo github.Repo, pulls github.Pulls) error {
	// Define the combined branch name
	workingBranchName := combineBranchName + workingBranchSuffix

	// Get the default branch of the repository
	repoDefaultBranch, err := getDefaultBranch(ctx, restClient, repo.Owner, repo.Repo)
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}

	baseBranchSHA, err := getBranchSHA(ctx, restClient, repo.Owner, repo.Repo, repoDefaultBranch)
	if err != nil {
		return fmt.Errorf("failed to get SHA of main branch: %w", err)
	}

	// Delete any pre-existing working branch
	err = deleteBranch(ctx, restClient, repo.Owner, repo.Repo, workingBranchName)
	if err != nil {
		Logger.Debug("Working branch not found, continuing", "branch", workingBranchName)
	}

	// Delete any pre-existing combined branch
	err = deleteBranch(ctx, restClient, repo.Owner, repo.Repo, combineBranchName)
	if err != nil {
		Logger.Debug("Combined branch not found, continuing", "branch", combineBranchName)
	}

	// Create the combined branch
	err = createBranch(ctx, restClient, repo.Owner, repo.Repo, combineBranchName, baseBranchSHA)
	if err != nil {
		return fmt.Errorf("failed to create combined branch: %w", err)
	}

	// Create the working branch
	err = createBranch(ctx, restClient, repo.Owner, repo.Repo, workingBranchName, baseBranchSHA)
	if err != nil {
		return fmt.Errorf("failed to create working branch: %w", err)
	}

	// Merge all PR branches into the working branch
	var combinedPRs []string
	var mergeFailedPRs []string
	for _, pr := range pulls {
		err := mergeBranch(ctx, restClient, repo.Owner, repo.Repo, workingBranchName, pr.Head.Ref)
		if err != nil {
			// Check if the error is a 409 merge conflict
			if isMergeConflictError(err) {
				// Log merge conflicts at DEBUG level
				Logger.Debug("Merge conflict", "branch", pr.Head.Ref, "error", err)
			} else {
				// Log other errors at WARN level
				Logger.Warn("Failed to merge branch", "branch", pr.Head.Ref, "error", err)
			}
			mergeFailedPRs = append(mergeFailedPRs, fmt.Sprintf("#%d", pr.Number))
		} else {
			Logger.Debug("Merged branch", "branch", pr.Head.Ref)
			combinedPRs = append(combinedPRs, fmt.Sprintf("#%d - %s", pr.Number, pr.Title))
		}
	}

	// Update the combined branch to the latest commit of the working branch
	err = updateRef(ctx, restClient, repo.Owner, repo.Repo, combineBranchName, workingBranchName)
	if err != nil {
		return fmt.Errorf("failed to update combined branch: %w", err)
	}

	// Delete the temporary working branch
	err = deleteBranch(ctx, restClient, repo.Owner, repo.Repo, workingBranchName)
	if err != nil {
		Logger.Warn("Failed to delete working branch", "branch", workingBranchName, "error", err)
	}

	// Create the combined PR
	prBody := generatePRBody(combinedPRs, mergeFailedPRs)
	prTitle := "Combined PRs"
	err = createPullRequest(ctx, restClient, repo.Owner, repo.Repo, prTitle, combineBranchName, repoDefaultBranch, prBody)
	if err != nil {
		return fmt.Errorf("failed to create combined PR: %w", err)
	}

	return nil
}

// isMergeConflictError checks if the error is a 409 Merge Conflict
func isMergeConflictError(err error) bool {
	// Check if the error message contains "HTTP 409: Merge conflict"
	return err != nil && strings.Contains(err.Error(), "HTTP 409: Merge conflict")
}

// Find the default branch of a repository
func getDefaultBranch(ctx context.Context, client *api.RESTClient, owner, repo string) (string, error) {
	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	endpoint := fmt.Sprintf("repos/%s/%s", owner, repo)
	err := client.Get(endpoint, &repoInfo)
	if err != nil {
		return "", fmt.Errorf("failed to get default branch: %w", err)
	}
	return repoInfo.DefaultBranch, nil
}

// Get the SHA of a given branch
func getBranchSHA(ctx context.Context, client *api.RESTClient, owner, repo, branch string) (string, error) {
	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	endpoint := fmt.Sprintf("repos/%s/%s/git/ref/heads/%s", owner, repo, branch)
	err := client.Get(endpoint, &ref)
	if err != nil {
		return "", fmt.Errorf("failed to get SHA of branch %s: %w", branch, err)
	}
	return ref.Object.SHA, nil
}

// generatePRBody generates the body for the combined PR
func generatePRBody(combinedPRs, mergeFailedPRs []string) string {
	body := "✅ The following pull requests have been successfully combined:\n"
	for _, pr := range combinedPRs {
		body += "- " + pr + "\n"
	}
	if len(mergeFailedPRs) > 0 {
		body += "\n⚠️ The following pull requests could not be merged due to conflicts:\n"
		for _, pr := range mergeFailedPRs {
			body += "- " + pr + "\n"
		}
	}

	body += "\n> Generated with [gh-combine](https://github.com/github/gh-combine)"

	return body
}

// deleteBranch deletes a branch in the repository
func deleteBranch(ctx context.Context, client *api.RESTClient, owner, repo, branch string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
	return client.Delete(endpoint, nil)
}

// createBranch creates a new branch in the repository
func createBranch(ctx context.Context, client *api.RESTClient, owner, repo, branch, sha string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/git/refs", owner, repo)
	payload := map[string]string{
		"ref": "refs/heads/" + branch,
		"sha": sha,
	}
	body, err := encodePayload(payload)
	if err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}
	return client.Post(endpoint, body, nil)
}

// mergeBranch merges a branch into the base branch
func mergeBranch(ctx context.Context, client *api.RESTClient, owner, repo, base, head string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/merges", owner, repo)
	payload := map[string]string{
		"base": base,
		"head": head,
	}
	body, err := encodePayload(payload)
	if err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}
	return client.Post(endpoint, body, nil)
}

// updateRef updates a branch to point to the latest commit of another branch
func updateRef(ctx context.Context, client *api.RESTClient, owner, repo, branch, sourceBranch string) error {
	// Get the SHA of the source branch
	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	endpoint := fmt.Sprintf("repos/%s/%s/git/ref/heads/%s", owner, repo, sourceBranch)
	err := client.Get(endpoint, &ref)
	if err != nil {
		return fmt.Errorf("failed to get SHA of source branch: %w", err)
	}

	// Update the branch to point to the new SHA
	endpoint = fmt.Sprintf("repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
	payload := map[string]interface{}{
		"sha":   ref.Object.SHA,
		"force": true,
	}
	body, err := encodePayload(payload)
	if err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}
	return client.Patch(endpoint, body, nil)
}

// createPullRequest creates a new pull request
func createPullRequest(ctx context.Context, client *api.RESTClient, owner, repo, title, head, base, body string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/pulls", owner, repo)
	payload := map[string]string{
		"title": title,
		"head":  head,
		"base":  base,
		"body":  body,
	}
	requestBody, err := encodePayload(payload)
	if err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}
	return client.Post(endpoint, requestBody, nil)
}

// encodePayload encodes a payload as JSON and returns an io.Reader
func encodePayload(payload interface{}) (io.Reader, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
