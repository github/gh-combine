package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/github/gh-combine/internal/github"
)

var (
	ErrEmptyRepositoriesFilePath = fmt.Errorf("empty repositories file path")
)

func ParseRepositories(args []string, path string) ([]github.Repo, error) {
	if len(args) == 0 && reposFile == "" {
		return nil, nil
	}

	argsRepos, err := parseRepositoriesArgs(args)
	if err != nil {
		return nil, err
	}

	fileRepos, err := parseRepositoriesFile(path)
	if err != nil {
		return nil, err
	}

	return append(argsRepos, fileRepos...), nil
}

func parseRepositoriesArgs(args []string) ([]github.Repo, error) {
	repos := []github.Repo{}

	for _, arg := range args {
		for _, rawRepo := range strings.Split(arg, ",") {

			repo, err := github.ParseRepo(rawRepo)
			if err != nil {
				return nil, err
			}

			repos = append(repos, repo)
		}
	}

	return repos, nil
}

// TODO: this should be removed to accept `gh-combine < repos` instead.
func parseRepositoriesFile(path string) ([]github.Repo, error) {
	if path == "" {
		return nil, nil
	}

	repos := []github.Repo{}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read repositories file %s: %w", path, err)
	}

	lines := strings.Split(string(fileContent), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		repo, err := github.ParseRepo(line)
		if err != nil {
			return nil, err
		}

		repos = append(repos, repo)
	}

	return repos, nil
}
