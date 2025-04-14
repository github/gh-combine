package cmd

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/github/gh-combine/internal/common"
)

// checks if a PR matches all filtering criteria
func PrMatchesCriteria(branch string, prLabels []string) bool {
	// Check branch criteria if any are specified
	if !branchMatchesCriteria(branch) {
		return false
	}

	// Check label criteria if any are specified
	if !labelsMatch(prLabels, ignoreLabels, selectLabels, caseSensitiveLabels) {
		return false
	}

	return true
}

// checks if a branch matches the branch filtering criteria
func branchMatchesCriteria(branch string) bool {
	Logger.Debug("Checking branch criteria", "branch", branch)
	// Do not attempt to match on existing branches that were created by this CLI
	if branch == combineBranchName {
		Logger.Debug("Branch is a combine branch, skipping match")
		return false
	}

	// If no branch filters are specified, all branches pass this check
	if branchPrefix == "" && branchSuffix == "" && branchRegex == "" {
		Logger.Debug("No branch filters specified, passing match")
		return true
	}

	// Apply branch prefix filter if specified
	if branchPrefix != "" && !strings.HasPrefix(branch, branchPrefix) {
		Logger.Debug("Branch does not match prefix", "prefix", branchPrefix, "branch", branch)
		return false
	}

	// Apply branch suffix filter if specified
	if branchSuffix != "" && !strings.HasSuffix(branch, branchSuffix) {
		Logger.Debug("Branch does not match suffix", "suffix", branchSuffix, "branch", branch)
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
			Logger.Debug("Branch does not match regex", "regex", branchRegex, "branch", branch)
			return false
		}
	}

	Logger.Debug("Branch matches all branch criteria", "branch", branch)
	return true
}

func labelsMatch(prLabels, ignoreLabels, selectLabels []string, caseSensitive bool) bool {
	// If no ignoreLabels or selectLabels are specified, all labels pass this check
	if len(ignoreLabels) == 0 && len(selectLabels) == 0 {
		return true
	}

	// Normalize labels for case-insensitive matching if caseSensitive is false
	if !caseSensitive {
		prLabels = common.NormalizeArray(prLabels)
		ignoreLabels = common.NormalizeArray(ignoreLabels)
		selectLabels = common.NormalizeArray(selectLabels)
	}

	// If the pull request contains any of the ignore labels, it doesn't match
	for _, l := range ignoreLabels {
		if slices.Contains(prLabels, l) {
			return false
		}
	}

	// If selectLabels are specified but the pull request has no labels, it doesn't match
	if len(selectLabels) > 0 && len(prLabels) == 0 {
		return false
	}

	// If the pull request contains any of the select labels, it matches
	for _, l := range selectLabels {
		if slices.Contains(prLabels, l) {
			return true
		}
	}

	// If none of the select labels are found, it doesn't match
	return len(selectLabels) == 0
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

// PrMeetsRequirements checks if a PR meets additional requirements beyond basic criteria
func PrMeetsRequirements(ctx context.Context, graphQlClient *api.GraphQLClient, owner, repo string, prNumber int) (bool, error) {
	// If no additional requirements are specified, the PR meets requirements
	if !requireCI && !mustBeApproved {
		return true, nil
	}

	// Fetch PR status info once
	response, err := GetPRStatusInfo(ctx, graphQlClient, owner, repo, prNumber)
	if err != nil {
		return false, err
	}

	// Check CI status if required
	if requireCI {
		passing := isCIPassing(response)
		if !passing {
			return false, nil
		}
	}

	// Check approval status if required
	if mustBeApproved {
		approved := isPRApproved(response)
		if !approved {
			return false, nil
		}
	}

	return true, nil
}

// isCIPassing checks if the CI status is passing based on the response
func isCIPassing(response *prStatusResponse) bool {
	commits := response.Data.Repository.PullRequest.Commits.Nodes
	if len(commits) == 0 {
		Logger.Debug("No commits found for PR")
		return false
	}

	statusCheckRollup := commits[0].Commit.StatusCheckRollup
	if statusCheckRollup == nil {
		Logger.Debug("No status checks found for PR")
		return true // If no checks defined, consider it passing
	}

	if statusCheckRollup.State != "SUCCESS" {
		Logger.Debug("PR failed CI check", "status", statusCheckRollup.State)
		return false
	}

	return true
}

// isPRApproved checks if the PR is approved based on the response
func isPRApproved(response *prStatusResponse) bool {
	reviewDecision := response.Data.Repository.PullRequest.ReviewDecision
	Logger.Debug("PR review decision", "decision", reviewDecision)

	switch reviewDecision {
	case "APPROVED":
		return true
	case "": // When no reviews are required
		Logger.Debug("PR has no required reviewers")
		return true // If no reviews required, consider it approved
	default:
		Logger.Debug("PR not approved", "decision", reviewDecision)
		return false
	}
}
