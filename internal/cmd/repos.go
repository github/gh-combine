package cmd

import (
	"fmt"
	"os"
	"strings"
)

// ParseRepositories parses repository names from arguments or file with support for default owner
func ParseRepositories(args []string, reposFile string, defaultOwner string) ([]string, error) {
	// Explicitly initialize repos as an empty slice
	repos := []string{}

	// If both args and reposFile are empty, return an empty slice
	if len(args) == 0 && reposFile == "" {
		return repos, nil
	}

	// Parse from command line arguments
	if len(args) > 0 {
		// Check if repos are comma-separated
		for _, arg := range args {
			if strings.Contains(arg, ",") {
				splitRepos := strings.Split(arg, ",")
				for _, repo := range splitRepos {
					if trimmedRepo := strings.TrimSpace(repo); trimmedRepo != "" {
						repos = append(repos, applyDefaultOwner(trimmedRepo, defaultOwner))
					}
				}
			} else {
				repos = append(repos, applyDefaultOwner(arg, defaultOwner))
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
			// Trim whitespace and ignore comments
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
				continue
			}

			// Remove inline comments
			if idx := strings.Index(trimmedLine, "#"); idx != -1 {
				trimmedLine = strings.TrimSpace(trimmedLine[:idx])
			}

			if trimmedLine != "" {
				repos = append(repos, applyDefaultOwner(trimmedLine, defaultOwner))
			}
		}
	}

	return repos, nil
}

// applyDefaultOwner adds the default owner to a repo name if it doesn't already have an owner
func applyDefaultOwner(repo string, defaultOwner string) string {
	if defaultOwner == "" || strings.Contains(repo, "/") {
		return repo
	}
	return defaultOwner + "/" + repo
}
