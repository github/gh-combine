package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

// checks if a PR matches all filtering criteria
func PrMatchesCriteria(branch string, prLabels []struct{ Name string }) bool {
	// Check branch criteria if any are specified
	if !branchMatchesCriteria(branch) {
		return false
	}

	// Check label criteria if any are specified
	if !labelsMatchCriteria(prLabels) {
		return false
	}

	return true
}

// checks if a branch matches the branch filtering criteria
func branchMatchesCriteria(branch string) bool {
	// If no branch filters are specified, all branches pass this check
	if branchPrefix == "" && branchSuffix == "" && branchRegex == "" {
		return true
	}

	// Apply branch prefix filter if specified
	if branchPrefix != "" && !strings.HasPrefix(branch, branchPrefix) {
		return false
	}

	// Apply branch suffix filter if specified
	if branchSuffix != "" && !strings.HasSuffix(branch, branchSuffix) {
		return false
	}

	// Apply branch regex filter if specified
	if branchRegex != "" {
		regex, err := regexp.Compile(branchRegex)
		if err != nil {
			Logger.Warn("Invalid regex pattern", "pattern", branchRegex, "error", err)
			return false
		}

		if !regex.MatchString(branch) {
			return false
		}
	}

	return true
}

// labelsMatchCriteria checks if PR labels match the label filtering criteria
func labelsMatchCriteria(prLabels []struct{ Name string }) bool {
	// If no label filters are specified, all PRs pass this check
	if ignoreLabel == "" && len(ignoreLabels) == 0 &&
		selectLabel == "" && len(selectLabels) == 0 {
		return true
	}

	// Check for ignore label (singular)
	if ignoreLabel != "" {
		for _, label := range prLabels {
			if label.Name == ignoreLabel {
				return false
			}
		}
	}

	// Check for ignore labels (plural)
	if len(ignoreLabels) > 0 {
		for _, ignoreL := range ignoreLabels {
			for _, prLabel := range prLabels {
				if prLabel.Name == ignoreL {
					return false
				}
			}
		}
	}

	// Check for select label (singular)
	if selectLabel != "" {
		found := false
		for _, label := range prLabels {
			if label.Name == selectLabel {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check for select labels (plural)
	if len(selectLabels) > 0 {
		for _, requiredLabel := range selectLabels {
			found := false
			for _, prLabel := range prLabels {
				if prLabel.Name == requiredLabel {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// GraphQL response structure for PR status info
type prStatusResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				ReviewDecision string `json:"reviewDecision"`
				Commits        struct {
					Nodes []struct {
						Commit struct {
							StatusCheckRollup *struct {
								State string `json:"state"`
							} `json:"statusCheckRollup"`
						} `json:"commit"`
					} `json:"nodes"`
				} `json:"commits"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// GetPRStatusInfo fetches both CI status and approval status using GitHub's GraphQL API
func GetPRStatusInfo(ctx context.Context, graphQlClient *api.GraphQLClient, owner, repo string, prNumber int) (*prStatusResponse, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Continue processing
	}

	// Define a struct with embedded graphql query
	var query struct {
		Repository struct {
			PullRequest struct {
				ReviewDecision string
				Commits        struct {
					Nodes []struct {
						Commit struct {
							StatusCheckRollup *struct {
								State string
							}
						}
					}
				} `graphql:"commits(last: 1)"`
			} `graphql:"pullRequest(number: $prNumber)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	// Prepare GraphQL query variables
	variables := map[string]interface{}{
		"owner":    graphql.String(owner),
		"repo":     graphql.String(repo),
		"prNumber": graphql.Int(prNumber),
	}

	// Execute GraphQL query
	err := graphQlClient.Query("PullRequestStatus", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("GraphQL query failed: %w", err)
	}

	// Convert to our response format
	response := &prStatusResponse{}
	response.Data.Repository.PullRequest.ReviewDecision = query.Repository.PullRequest.ReviewDecision

	if len(query.Repository.PullRequest.Commits.Nodes) > 0 {
		response.Data.Repository.PullRequest.Commits.Nodes = make([]struct {
			Commit struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		}, len(query.Repository.PullRequest.Commits.Nodes))

		for i, node := range query.Repository.PullRequest.Commits.Nodes {
			if node.Commit.StatusCheckRollup != nil {
				response.Data.Repository.PullRequest.Commits.Nodes[i].Commit.StatusCheckRollup = &struct {
					State string `json:"state"`
				}{
					State: node.Commit.StatusCheckRollup.State,
				}
			}
		}
	}

	return response, nil
}

// HasPassingCI checks if a pull request has passing CI
func HasPassingCI(ctx context.Context, graphQlClient *api.GraphQLClient, owner, repo string, prNumber int) (bool, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		// Continue processing
	}

	// Get PR status info using GraphQL
	response, err := GetPRStatusInfo(ctx, graphQlClient, owner, repo, prNumber)
	if err != nil {
		return false, err
	}

	// Get the commit status check info
	commits := response.Data.Repository.PullRequest.Commits.Nodes
	if len(commits) == 0 {
		Logger.Debug("No commits found for PR", "repo", fmt.Sprintf("%s/%s", owner, repo), "pr", prNumber)
		return false, nil
	}

	// Get status check info
	statusCheckRollup := commits[0].Commit.StatusCheckRollup
	if statusCheckRollup == nil {
		Logger.Debug("No status checks found for PR", "repo", fmt.Sprintf("%s/%s", owner, repo), "pr", prNumber)
		return true, nil // If no checks defined, consider it passing
	}

	// Check if status is SUCCESS
	if statusCheckRollup.State != "SUCCESS" {
		Logger.Debug("PR failed CI check", "repo", fmt.Sprintf("%s/%s", owner, repo), "pr", prNumber, "status", statusCheckRollup.State)
		return false, nil
	}

	return true, nil
}

// HasApproval checks if a pull request has been approved
func HasApproval(ctx context.Context, graphQlClient *api.GraphQLClient, owner, repo string, prNumber int) (bool, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		// Continue processing
	}

	// Get PR status info using GraphQL
	response, err := GetPRStatusInfo(ctx, graphQlClient, owner, repo, prNumber)
	if err != nil {
		return false, err
	}

	reviewDecision := response.Data.Repository.PullRequest.ReviewDecision
	Logger.Debug("PR review decision", "repo", fmt.Sprintf("%s/%s", owner, repo), "pr", prNumber, "decision", reviewDecision)

	// Check the review decision
	switch reviewDecision {
	case "APPROVED":
		return true, nil
	case "": // When no reviews are required
		Logger.Debug("PR has no required reviewers", "repo", fmt.Sprintf("%s/%s", owner, repo), "pr", prNumber)
		return true, nil // If no reviews required, consider it approved
	default:
		// Any other decision (REVIEW_REQUIRED, CHANGES_REQUESTED, etc.)
		Logger.Debug("PR not approved", "repo", fmt.Sprintf("%s/%s", owner, repo), "pr", prNumber, "decision", reviewDecision)
		return false, nil
	}
}

// PrMeetsRequirements checks if a PR meets additional requirements beyond basic criteria
func PrMeetsRequirements(ctx context.Context, graphQlClient *api.GraphQLClient, owner, repo string, prNumber int) (bool, error) {
	// If no additional requirements are specified, the PR meets requirements
	if !requireCI && !mustBeApproved {
		return true, nil
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		// Continue processing
	}

	// Check CI status if required
	if requireCI {
		passing, err := HasPassingCI(ctx, graphQlClient, owner, repo, prNumber)
		if err != nil {
			return false, err
		}
		if !passing {
			return false, nil
		}
	}

	// Check approval status if required
	if mustBeApproved {
		approved, err := HasApproval(ctx, graphQlClient, owner, repo, prNumber)
		if err != nil {
			return false, err
		}
		if !approved {
			return false, nil
		}
	}

	return true, nil
}
