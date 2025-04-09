package cmd

import (
	"errors"
	"fmt"
)

// validateInputs checks if the provided inputs are valid
func ValidateInputs(args []string) error {
	// Check that ignore-label and select-label are not the same
	if ignoreLabel != "" && selectLabel != "" && ignoreLabel == selectLabel {
		return errors.New("--ignore-label(s) and --label(s) cannot have the same value")
	}

	// Check for conflicts between ignoreLabels and selectLabel
	if selectLabel != "" && len(ignoreLabels) > 0 {
		for _, ignoreL := range ignoreLabels {
			if ignoreL == selectLabel {
				return fmt.Errorf("--ignore-labels contains %q which conflicts with --label %q", ignoreL, selectLabel)
			}
		}
	}

	// Check for conflicts between ignoreLabel and selectLabels
	if ignoreLabel != "" && len(selectLabels) > 0 {
		for _, selectL := range selectLabels {
			if selectL == ignoreLabel {
				return fmt.Errorf("--label(s) contains %q which conflicts with --ignore-label %q", selectL, ignoreLabel)
			}
		}
	}

	// Check for conflicts between ignoreLabels and selectLabels
	if len(ignoreLabels) > 0 && len(selectLabels) > 0 {
		for _, ignoreL := range ignoreLabels {
			for _, selectL := range selectLabels {
				if ignoreL == selectL {
					return fmt.Errorf("--ignore-labels contains %q which conflicts with --labels containing the same value", ignoreL)
				}
			}
		}
	}

	// If no args and no file, we can't proceed
	if len(args) == 0 && reposFile == "" {
		return errors.New("must specify repositories or provide a file containing a list of repositories with --file")
	}

	// Warn if no filtering options are provided at all
	if branchPrefix == "" && branchSuffix == "" && branchRegex == "" &&
		ignoreLabel == "" && len(ignoreLabels) == 0 && selectLabel == "" && len(selectLabels) == 0 &&
		!requireCI && !mustBeApproved {
		Logger.Warn("No filtering options specified. This will attempt to combine ALL open pull requests. Use --label, --labels, --ignore-label, --ignore-labels, --branch-prefix, --branch-suffix, --branch-regex, --dependabot, etc to filter.")
	}

	return nil
}
