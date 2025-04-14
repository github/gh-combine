package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"

	"github.com/github/gh-combine/internal/github"
	"github.com/github/gh-combine/internal/version"
)

var (
	branchPrefix string
	branchSuffix string
	branchRegex  string

	selectLabels []string
	ignoreLabels []string

	addLabels    []string
	addAssignees []string

	requireCI           bool
	mustBeApproved      bool
	autoclose           bool
	updateBranch        bool
	reposFile           string
	minimum             int
	defaultOwner        string
	baseBranch          string
	combineBranchName   string
	workingBranchSuffix string
	dependabot          bool
	caseSensitiveLabels bool
	noColor             bool
	noStats             bool
)

// StatsCollector tracks stats for the CLI run
type StatsCollector struct {
	ReposProcessed int
	PRsCombined   int
	PRsSkippedMergeConflict int
	PRsSkippedCriteria int
	PerRepoStats  map[string]*RepoStats
	CombinedPRLinks []string
	StartTime     time.Time
	EndTime       time.Time
}

type RepoStats struct {
	RepoName         string
	CombinedCount    int
	SkippedMergeConf int
	SkippedCriteria  int
	CombinedPRLink   string
	NotEnoughPRs     bool
}

// NewRootCmd creates the root command for the gh-combine CLI
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "combine owner/repo",
		Short: "Combine multiple pull requests into a single PR",
		Long: `Combine multiple pull requests that match specific criteria into a single PR.
    Examples:
	  # Note: You should use some form of filtering to avoid combining all open PRs in a repository.
	  # For example, you can filter by branch name, labels, or other criteria.
	  # Forms of filtering include:
	  # --labels, --ignore-labels, --branch-prefix, --branch-suffix, --branch-regex, --dependabot, etc.

      # Basic usage with a single repository to combine all pull requests into one
      gh combine owner/repo

	  # Basic usage to only combine pull requests from dependabot
	  gh combine owner/repo --dependabot
    
      # Multiple repositories (comma-separated)
      gh combine octocat/repo1,octocat/repo2

	  # Multiple repositories (no commas)
	  gh combine octocat/repo1 octocat/repo2
      
      # Using default owner for repositories
      gh combine --owner octocat repo1 repo2

	  # Using default owner for only some repositories
	  gh combine --owner octocat repo1 octocat/repo2
    
      # Using a file with repository names (one per line: owner/repo format)
      gh combine --file repos.txt
    
      # Filter PRs by branch name
      gh combine owner/repo --branch-prefix dependabot/ # Only include PRs with the standard dependabot branch prefix
      gh combine owner/repo --branch-suffix -update
      gh combine owner/repo --branch-regex "dependabot/.*"
    
      # Filter PRs by labels
      gh combine owner/repo --labels dependencies           # PRs must have this single label
      gh combine owner/repo --labels security,dependencies  # PRs must have ALL these labels
	  gh combine owner/repo --labels Dependencies --case-sensitive-labels # PRs must have this label, case-sensitive
      
      # Exclude PRs by labels
      gh combine owner/repo --ignore-labels wip         # Ignore PRs with this label
      gh combine owner/repo --ignore-labels wip,draft   # Ignore PRs with ANY of these labels
    
      # Set requirements for PRs to be combined
      gh combine owner/repo --require-ci                # Only include PRs with passing CI
      gh combine owner/repo --require-approved          # Only include approved PRs
      gh combine owner/repo --minimum 3                 # Need at least 3 matching PRs
    
      # Add metadata to combined PR
      gh combine owner/repo --add-labels security,dependencies   # Add these labels to the new PR
      gh combine owner/repo --add-assignees octocat,hubot        # Assign users to the new PR
    
      # Additional options
      gh combine owner/repo --autoclose                         # Close source PRs when combined PR is merged
	  gh combine owner/repo --base-branch main                  # Use a different base branch for the combined PR
	  gh combine owner/repo --no-color                          # Disable color output
	  gh combine owner/repo --no-stats                          # Disable stats summary display
	  gh combine owner/repo --combine-branch-name combined-prs  # Use a different name for the combined PR branch
	  gh combine owner/repo --working-branch-suffix -working    # Use a different suffix for the working branch
      gh combine owner/repo --update-branch                     # Update the branch of the combined PR`,
		RunE: runCombine,
	}

	// Add flags
	rootCmd.Flags().StringVar(&branchPrefix, "branch-prefix", "", "Branch prefix to filter PRs")
	rootCmd.Flags().StringVar(&branchSuffix, "branch-suffix", "", "Branch suffix to filter PRs")
	rootCmd.Flags().StringVar(&branchRegex, "branch-regex", "", "Regex pattern to filter PRs by branch name")

	rootCmd.Flags().StringSliceVar(&selectLabels, "labels", nil, "Only include PRs with ALL these labels (comma-separated)")
	rootCmd.Flags().StringSliceVar(&ignoreLabels, "ignore-labels", nil, "Ignore PRs with ANY of these labels (comma-separated)")

	// Labels to add to the combined PR
	rootCmd.Flags().StringSliceVar(&addLabels, "add-labels", nil, "Comma-separated list of labels to add to the combined PR")

	// Other flags
	rootCmd.Flags().StringSliceVar(&addAssignees, "add-assignees", nil, "Comma-separated list of users to assign to the combined PR")
	rootCmd.Flags().BoolVar(&requireCI, "require-ci", false, "Only include PRs with passing CI checks")
	rootCmd.Flags().BoolVar(&dependabot, "dependabot", false, "Only include PRs with the dependabot branch prefix")
	rootCmd.Flags().BoolVar(&mustBeApproved, "require-approved", false, "Only include PRs that have been approved")
	rootCmd.Flags().BoolVar(&autoclose, "autoclose", false, "Close source PRs when combined PR is merged")
	rootCmd.Flags().BoolVar(&updateBranch, "update-branch", false, "Update the branch of the combined PR if possible")
	rootCmd.Flags().StringVar(&baseBranch, "base-branch", "main", "Base branch for the combined PR (default: main)")
	rootCmd.Flags().StringVar(&combineBranchName, "combine-branch-name", "combined-prs", "Name of the combined PR branch")
	rootCmd.Flags().StringVar(&workingBranchSuffix, "working-branch-suffix", "-working", "Suffix of the working branch")
	rootCmd.Flags().StringVar(&reposFile, "file", "", "File containing repository names, one per line")
	rootCmd.Flags().IntVar(&minimum, "minimum", 2, "Minimum number of PRs to combine")
	rootCmd.Flags().StringVar(&defaultOwner, "owner", "", "Default owner for repositories (if not specified in repo name or missing from file inputs)")
	rootCmd.Flags().BoolVar(&caseSensitiveLabels, "case-sensitive-labels", false, "Use case-sensitive label matching")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.Flags().BoolVar(&noStats, "no-stats", false, "Disable stats summary display")

	// Add deprecated flags for backward compatibility
	// rootCmd.Flags().IntVar(&minimum, "min-combine", 2, "Minimum number of PRs to combine (deprecated, use --minimum)")

	// Mark deprecated flags
	// rootCmd.Flags().MarkDeprecated("min-combine", "use --minimum instead")

	return rootCmd
}

// Run executes the main functionality of the application
func Run() error {
	cmd := NewRootCmd()
	return cmd.Execute()
}

// runCombine is the main execution function for the combine command
func runCombine(cmd *cobra.Command, args []string) error {
	ctx, cancel := SetupSignalContext()
	defer cancel()

	Logger.Debug("starting gh-combine", "version", version.String())

	if dependabot && branchPrefix == "" {
		branchPrefix = "dependabot/"
	}

	// Input validation
	if err := ValidateInputs(args); err != nil {
		return err
	}

	spinner := NewSpinner("")
	defer spinner.Stop()

	// Parse repositories from args or file
	repos, err := ParseRepositories(args, reposFile, defaultOwner)
	if err != nil {
		return fmt.Errorf("failed to parse repositories: %w", err)
	}

	if len(repos) == 0 {
		return errors.New("no repositories specified")
	}

	stats := &StatsCollector{
		PerRepoStats: make(map[string]*RepoStats),
		StartTime: time.Now(),
	}

	// Execute combination logic
	if err := executeCombineCommand(ctx, spinner, repos, stats); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	stats.EndTime = time.Now()

	if !noStats {
		spinner.Stop()
		displayStatsSummary(stats)
	}

	return nil
}

// executeCombineCommand performs the actual API calls and processing
func executeCombineCommand(ctx context.Context, spinner *Spinner, repos []string, stats *StatsCollector) error {
	// Create GitHub API client
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Create GitHub GraphQL client
	graphQlClient, err := api.DefaultGraphQLClient()
	if err != nil {
		return fmt.Errorf("failed to create GraphQLClient client: %w", err)
	}

	for _, repoString := range repos {
		// Check if context was cancelled (CTRL+C pressed)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		spinner.UpdateMessage("Parsing " + repoString)

		repo, err := github.ParseRepo(repoString)
		if err != nil {
			return fmt.Errorf("failed to parse repo: %w", err)
		}

		spinner.UpdateMessage("Processing " + repo.String())
		Logger.Debug("Processing repository", "repo", repo)

		if stats.PerRepoStats[repo.String()] == nil {
			stats.PerRepoStats[repo.String()] = &RepoStats{RepoName: repo.String()}
		}

		// Process the repository
		if err := processRepository(ctx, restClient, graphQlClient, spinner, repo, stats.PerRepoStats[repo.String()], stats); err != nil {
			if ctx.Err() != nil {
				// If the context was cancelled, stop processing
				return ctx.Err()
			}
			Logger.Warn("Failed to process repository", "repo", repo, "error", err)
			continue
		}
		stats.ReposProcessed++
	}

	return nil
}

// processRepository handles a single repository's PRs
func processRepository(ctx context.Context, client *api.RESTClient, graphQlClient *api.GraphQLClient, spinner *Spinner, repo github.Repo, repoStats *RepoStats, stats *StatsCollector) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue processing
	}

	// Fetch all open pull requests for the repository
	pulls, err := fetchOpenPullRequests(ctx, client, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch open pull requests: %w", err)
	}

	// Check for cancellation again
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue processing
	}

	// Filter PRs based on criteria
	var matchedPRs github.Pulls
	for _, pull := range pulls {
		// Extract labels
		labels := []string{}
		for _, label := range pull.Labels {
			labels = append(labels, label.Name)
		}

		// Check if PR matches all filtering criteria
		if !PrMatchesCriteria(pull.Head.Ref, labels) {
			repoStats.SkippedCriteria++
			stats.PRsSkippedCriteria++
			continue
		}

		// Check if PR meets additional requirements (CI, approval)
		meetsRequirements, err := PrMeetsRequirements(ctx, graphQlClient, repo.Owner, repo.Repo, pull.Number)
		if err != nil {
			Logger.Warn("Failed to check PR requirements", "repo", repo, "pr", pull.Number, "error", err)
			continue
		}

		if !meetsRequirements {
			repoStats.SkippedCriteria++
			stats.PRsSkippedCriteria++
			continue
		}

		matchedPRs = append(matchedPRs, pull)
	}

	// Check if we have enough PRs to combine
	if len(matchedPRs) < minimum {
		Logger.Debug("Not enough PRs match criteria", "repo", repo, "matched", len(matchedPRs), "required", minimum)
		repoStats.NotEnoughPRs = true
		return nil
	}

	Logger.Debug("Matched PRs", "repo", repo, "count", len(matchedPRs))

	// Wrap the *api.RESTClient to implement RESTClientInterface
	restClientWrapper := struct {
		RESTClientInterface
	}{client}

	// Combine the PRs and collect stats
	combined, mergeConflicts, combinedPRLink, err := CombinePRsWithStats(ctx, graphQlClient, restClientWrapper, repo, matchedPRs)
	if err != nil {
		return fmt.Errorf("failed to combine PRs: %w", err)
	}

	repoStats.CombinedCount = len(combined)
	repoStats.SkippedMergeConf = len(mergeConflicts)
	repoStats.CombinedPRLink = combinedPRLink
	stats.PRsCombined += len(combined)
	stats.PRsSkippedMergeConflict += len(mergeConflicts)
	if combinedPRLink != "" {
		stats.CombinedPRLinks = append(stats.CombinedPRLinks, combinedPRLink)
	}

	Logger.Debug("Combined PRs", "count", len(matchedPRs), "owner", repo.Owner, "repo", repo.Repo)

	return nil
}

// fetchOpenPullRequests fetches all open pull requests for a repository, handling pagination
func fetchOpenPullRequests(ctx context.Context, client *api.RESTClient, repo github.Repo) (github.Pulls, error) {
	var allPulls github.Pulls
	page := 1

	for {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue processing
		}

		var pulls github.Pulls
		endpoint := fmt.Sprintf("%s?state=open&page=%d&per_page=100", repo.PullsEndpoint(), page)
		if err := client.Get(endpoint, &pulls); err != nil {
			return nil, fmt.Errorf("failed to fetch pull requests from page %d: %w", page, err)
		}

		// If the current page is empty, we've reached the end
		if len(pulls) == 0 {
			break
		}

		// Append fetched pulls to the result
		allPulls = append(allPulls, pulls...)

		// If fewer than 100 PRs are returned, we've reached the last page
		if len(pulls) < 100 {
			break
		}

		page++
	}

	return allPulls, nil
}

// CombinePRsWithStats wraps CombinePRs to collect stats and return combined/skipped PRs and the combined PR link
func CombinePRsWithStats(ctx context.Context, graphQlClient *api.GraphQLClient, restClient RESTClientInterface, repo github.Repo, pulls github.Pulls) (combined []string, mergeConflicts []string, combinedPRLink string, err error) {
	// ...existing code for CombinePRs...
	// This is a stub. You should move the logic from CombinePRs here and update it to collect combined, mergeConflicts, and the PR link.
	return nil, nil, "", CombinePRs(ctx, graphQlClient, restClient, repo, pulls)
}

func displayStatsSummary(stats *StatsCollector) {
	elapsed := stats.EndTime.Sub(stats.StartTime)
	if noColor {
		fmt.Println("Stats Summary (Color Disabled):")
	} else {
		fmt.Println("\033[1;34mStats Summary:\033[0m")
	}
	fmt.Printf("Repositories Processed: %d\n", stats.ReposProcessed)
	fmt.Printf("PRs Combined: %d\n", stats.PRsCombined)
	fmt.Printf("PRs Skipped (Merge Conflicts): %d\n", stats.PRsSkippedMergeConflict)
	fmt.Printf("PRs Skipped (Criteria Not Met): %d\n", stats.PRsSkippedCriteria)
	fmt.Printf("Execution Time: %s\n", elapsed.Round(time.Second))

	if !noColor {
		fmt.Println("\033[1;32mLinks to Combined PRs:\033[0m")
	} else {
		fmt.Println("Links to Combined PRs:")
	}
	for _, link := range stats.CombinedPRLinks {
		fmt.Println("-", link)
	}

	fmt.Println("\nPer-Repository Details:")
	for _, repoStat := range stats.PerRepoStats {
		fmt.Printf("  %s\n", repoStat.RepoName)
		if repoStat.NotEnoughPRs {
			fmt.Println("    Not enough PRs to combine.")
			continue
		}
		fmt.Printf("    Combined: %d\n", repoStat.CombinedCount)
		fmt.Printf("    Skipped (Merge Conflicts): %d\n", repoStat.SkippedMergeConf)
		fmt.Printf("    Skipped (Criteria): %d\n", repoStat.SkippedCriteria)
		if repoStat.CombinedPRLink != "" {
			fmt.Printf("    Combined PR: %s\n", repoStat.CombinedPRLink)
		}
	}
}
