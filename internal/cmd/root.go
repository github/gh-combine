package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/spf13/cobra"

	"github.com/github/gh-combine/internal/version"
)

var (
	branchPrefix   string
	branchSuffix   string
	branchRegex    string
	selectLabel    string
	selectLabels   []string
	addLabels      []string
	addAssignees   []string
	requireCI      bool
	mustBeApproved bool
	autoclose      bool
	updateBranch   bool
	ignoreLabel    string
	ignoreLabels   []string
	reposFile      string
	minimum        int
	defaultOwner   string
)

// NewRootCmd creates the root command for the gh-combine CLI
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "combine owner/repo",
		Short: "Combine multiple pull requests into a single PR",
		Long: `Combine multiple pull requests that match specific criteria into a single PR.
    Examples:
      # Basic usage with a single repository (will default to "--branch-prefix dependabot/" and "--minimum 2")
      gh combine octocat/hello-world
    
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
      gh combine octocat/hello-world --branch-prefix dependabot/ # Only include PRs with the standard dependabot branch prefix
      gh combine octocat/hello-world --branch-suffix -update
      gh combine octocat/hello-world --branch-regex "dependabot/.*"
    
      # Filter PRs by labels
      gh combine octocat/hello-world --label dependencies        # PRs must have this single label
      gh combine octocat/hello-world --labels security,dependencies  # PRs must have ALL these labels
      
      # Exclude PRs by labels
      gh combine octocat/hello-world --ignore-label wip          # Ignore PRs with this label
      gh combine octocat/hello-world --ignore-labels wip,draft   # Ignore PRs with ANY of these labels
    
      # Set requirements for PRs to be combined
      gh combine octocat/hello-world --require-ci                # Only include PRs with passing CI
      gh combine octocat/hello-world --require-approved          # Only include approved PRs
      gh combine octocat/hello-world --minimum 3                 # Need at least 3 matching PRs
    
      # Add metadata to combined PR
      gh combine octocat/hello-world --add-labels security,dependencies   # Add these labels to the new PR
      gh combine octocat/hello-world --add-assignees octocat,hubot        # Assign users to the new PR
    
      # Additional options
      gh combine octocat/hello-world --autoclose                 # Close source PRs when combined PR is merged
      gh combine octocat/hello-world --update-branch             # Update the branch of the combined PR`,
		RunE: runCombine,
	}

	// Add flags
	rootCmd.Flags().StringVar(&branchPrefix, "branch-prefix", "dependabot/", "Branch prefix to filter PRs")
	rootCmd.Flags().StringVar(&branchSuffix, "branch-suffix", "", "Branch suffix to filter PRs")
	rootCmd.Flags().StringVar(&branchRegex, "branch-regex", "", "Regex pattern to filter PRs by branch name")

	// Label selection flags - singular and plural forms
	rootCmd.Flags().StringVar(&selectLabel, "label", "", "Only include PRs with this specific label")
	rootCmd.Flags().StringSliceVar(&selectLabels, "labels", nil, "Only include PRs with ALL these labels (comma-separated)")

	// Label ignoring flags - singular and plural forms
	rootCmd.Flags().StringVar(&ignoreLabel, "ignore-label", "", "Ignore PRs with this specific label")
	rootCmd.Flags().StringSliceVar(&ignoreLabels, "ignore-labels", nil, "Ignore PRs with ANY of these labels (comma-separated)")

	// Labels to add to the combined PR
	rootCmd.Flags().StringSliceVar(&addLabels, "add-labels", nil, "Comma-separated list of labels to add to the combined PR")

	// Other flags
	rootCmd.Flags().StringSliceVar(&addAssignees, "add-assignees", nil, "Comma-separated list of users to assign to the combined PR")
	rootCmd.Flags().BoolVar(&requireCI, "require-ci", false, "Only include PRs with passing CI checks")
	rootCmd.Flags().BoolVar(&mustBeApproved, "require-approved", false, "Only include PRs that have been approved")
	rootCmd.Flags().BoolVar(&autoclose, "autoclose", false, "Close source PRs when combined PR is merged")
	rootCmd.Flags().BoolVar(&updateBranch, "update-branch", false, "Update the branch of the combined PR if possible")
	rootCmd.Flags().StringVar(&reposFile, "file", "", "File containing repository names, one per line")
	rootCmd.Flags().IntVar(&minimum, "minimum", 2, "Minimum number of PRs to combine")
	rootCmd.Flags().StringVar(&defaultOwner, "owner", "", "Default owner for repositories (if not specified in repo name or missing from file inputs)")

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

	// Execute combination logic
	if err := executeCombineCommand(ctx, spinner, repos); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// executeCombineCommand performs the actual API calls and processing
func executeCombineCommand(ctx context.Context, spinner *Spinner, repos []string) error {
	// Create GitHub API client
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	for _, repo := range repos {
		// Check if context was cancelled (CTRL+C pressed)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		spinner.UpdateMessage("Processing " + repo)
		Logger.Debug("Processing repository", "repo", repo)

		// Process the repository
		if err := processRepository(ctx, restClient, spinner, repo); err != nil {
			if ctx.Err() != nil {
				// If the context was cancelled, stop processing
				return ctx.Err()
			}
			// Otherwise just log the error and continue
			Logger.Warn("Failed to process repository", "repo", repo, "error", err)
			continue
		}
	}

	return nil
}

// processRepository handles a single repository's PRs
func processRepository(ctx context.Context, client *api.RESTClient, spinner *Spinner, repo string) error {
	// Parse owner and repo name
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repo)
	}

	owner := parts[0]
	repoName := parts[1]

	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue processing
	}

	// Get open PRs for the repository
	var pulls []struct {
		Number int
		Title  string
		Head   struct {
			Ref string
		}
		Base struct {
			Ref string
			SHA string
		}
		Labels []struct {
			Name string
		}
	}

	endpoint := fmt.Sprintf("repos/%s/%s/pulls?state=open", owner, repoName)
	if err := client.Get(endpoint, &pulls); err != nil {
		return err
	}

	// Check for cancellation again
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue processing
	}

	// Filter PRs based on criteria
	var matchedPRs []struct {
		Number  int
		Title   string
		Branch  string
		Base    string
		BaseSHA string
	}

	for _, pull := range pulls {
		branch := pull.Head.Ref

		// Check if PR matches all filtering criteria
		if !PrMatchesCriteria(branch, pull.Labels) {
			continue
		}

		// TODO: Implement CI/approval status checking

		matchedPRs = append(matchedPRs, struct {
			Number  int
			Title   string
			Branch  string
			Base    string
			BaseSHA string
		}{
			Number:  pull.Number,
			Title:   pull.Title,
			Branch:  branch,
			Base:    pull.Base.Ref,
			BaseSHA: pull.Base.SHA,
		})
	}

	// Check if we have enough PRs to combine
	if len(matchedPRs) < minimum {
		Logger.Debug("Not enough PRs match criteria", "repo", repo, "matched", len(matchedPRs), "required", minimum)
		return nil
	}

	Logger.Debug("Matched PRs", "repo", repo, "count", len(matchedPRs))
	return nil
}
