package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

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
	assignees      []string
	requireCI      bool
	mustBeApproved bool
	autoclose      bool
	updateBranch   bool
	ignoreLabel    string
	ignoreLabels   []string
	reposFile      string
	minimum        int
)

// NewRootCmd creates the root command for the gh-combine CLI
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "combine [repo1,repo2,...]",
		Short: "Combine multiple pull requests into a single PR",
		Long: `Combine multiple pull requests that match specific criteria into a single PR.
	Examples:
	  # Basic usage with a single repository
	  gh combine octocat/hello-world
	
	  # Multiple repositories (comma-separated)
	  gh combine octocat/repo1,octocat/repo2
	
	  # Using a file with repository names (one per line)
	  gh combine --file repos.txt
	
	  # Filter PRs by branch name
	  gh combine octocat/hello-world --branch-prefix dependabot-
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
	  gh combine octocat/hello-world --assignees octocat,hubot            # Assign users to the new PR
	
	  # Additional options
	  gh combine octocat/hello-world --autoclose                 # Close source PRs when combined PR is merged
	  gh combine octocat/hello-world --update-branch             # Update the branch of the combined PR`,
		RunE: runCombine,
	}

	// Add flags
	rootCmd.Flags().StringVar(&branchPrefix, "branch-prefix", "", "Branch prefix to filter PRs")
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
	rootCmd.Flags().StringSliceVar(&assignees, "assignees", nil, "Comma-separated list of users to assign to the combined PR")
	rootCmd.Flags().BoolVar(&requireCI, "require-ci", false, "Only include PRs with passing CI checks")
	rootCmd.Flags().BoolVar(&mustBeApproved, "require-approved", false, "Only include PRs that have been approved")
	rootCmd.Flags().BoolVar(&autoclose, "autoclose", false, "Close source PRs when combined PR is merged")
	rootCmd.Flags().BoolVar(&updateBranch, "update-branch", false, "Update the branch of the combined PR if possible")
	rootCmd.Flags().StringVar(&reposFile, "file", "", "File containing repository names, one per line")
	rootCmd.Flags().IntVar(&minimum, "minimum", 2, "Minimum number of PRs to combine")

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
	ctx, cancel := setupSignalContext()
	defer cancel()

	Logger.Debug("starting gh-combine", "version", version.String())

	// Input validation
	if err := validateInputs(args); err != nil {
		return err
	}

	spinner := NewSpinner("")
	defer spinner.Stop()

	// Parse repositories from args or file
	repos, err := parseRepositories(args)
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

// setupSignalContext creates a context that's cancelled on SIGINT or SIGTERM
func setupSignalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-signalChan:
			Logger.Debug("Received interrupt signal, cancelling operations...")
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(signalChan)
	}()

	return ctx, cancel
}

// validateInputs checks if the provided inputs are valid
func validateInputs(args []string) error {
	// Check that ignore-label and select-label are not the same
	if ignoreLabel != "" && selectLabel != "" && ignoreLabel == selectLabel {
		return errors.New("--ignore-label and --select-label cannot have the same value")
	}

	// If no args and no file, we can't proceed
	if len(args) == 0 && reposFile == "" {
		return errors.New("must specify repositories or provide a file with --file")
	}

	// Warn if no filtering options are provided at all
	if branchPrefix == "" && branchSuffix == "" && branchRegex == "" &&
		ignoreLabel == "" && selectLabel == "" && len(selectLabels) == 0 &&
		!requireCI && !mustBeApproved {
		Logger.Warn("No filtering options specified. This will attempt to combine ALL open pull requests.")
	}

	return nil
}

// parseRepositories parses repository names from arguments or file
func parseRepositories(args []string) ([]string, error) {
	var repos []string

	// Parse from command line arguments
	if len(args) > 0 {
		// Check if repos are comma-separated
		for _, arg := range args {
			if strings.Contains(arg, ",") {
				splitRepos := strings.Split(arg, ",")
				for _, repo := range splitRepos {
					if trimmedRepo := strings.TrimSpace(repo); trimmedRepo != "" {
						repos = append(repos, trimmedRepo)
					}
				}
			} else {
				repos = append(repos, arg)
			}
		}
	}

	// Parse from file if specified
	if reposFile != "" {
		fileContent, err := os.ReadFile(reposFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read repositories file: %w", err)
		}

		lines := strings.Split(string(fileContent), "\n")
		for _, line := range lines {
			if trimmedLine := strings.TrimSpace(line); trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
				repos = append(repos, trimmedLine)
			}
		}
	}

	return repos, nil
}

// executeCombineCommand performs the actual API calls and processing
func executeCombineCommand(ctx context.Context, spinner *Spinner, repos []string) error {
	// Create GitHub API client
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	for _, repo := range repos {
		spinner.UpdateMessage("Processing " + repo)
		Logger.Debug("Processing repository", "repo", repo)

		// Parse owner and repo name
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			Logger.Warn("Invalid repository format, skipping", "repo", repo)
			continue
		}

		owner := parts[0]
		repoName := parts[1]

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
		if err := restClient.Get(endpoint, &pulls); err != nil {
			Logger.Warn("Failed to fetch PRs", "repo", repo, "error", err)
			continue
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
			if !prMatchesCriteria(branch, pull.Labels) {
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
			continue
		}

		// TODO: Implement PR combining logic
		// 1. Create combined branch
		// 2. Merge matched PR branches
		// 3. Create combined PR
		// 4. Add labels and assignees

		Logger.Debug("Matched PRs", "repo", repo, "count", len(matchedPRs))
	}

	return nil
}

// prMatchesCriteria checks if a PR matches all filtering criteria
func prMatchesCriteria(branch string, prLabels []struct{ Name string }) bool {
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

// branchMatchesCriteria checks if a branch matches the branch filtering criteria
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
