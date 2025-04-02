package cmd

import (
	"regexp"
	"strings"
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
