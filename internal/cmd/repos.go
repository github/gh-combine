package cmd

import (
	"fmt"
	"os"
	"strings"
)

// ParseRepositories parses repository names from arguments or file
func ParseRepositories(args []string, reposFile string) ([]string, error) {
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
