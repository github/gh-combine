package cmd

import (
	"errors"
	"fmt"
	"slices"
)

var errLabelsConflict = errors.New("--ignore-labels contains a value which conflicts with --labels")

// validateInputs checks if the provided inputs are valid
func ValidateInputs(args []string) error {
	if err := ValidateLabels(selectLabels, ignoreLabels); err != nil {
		return err
	}

	// If no args and no file, we can't proceed
	if len(args) == 0 && reposFile == "" {
		return errors.New("must specify repositories or provide a file containing a list of repositories with --file")
	}

	// Warn if no filtering options are provided at all
	if branchPrefix == "" && branchSuffix == "" && branchRegex == "" &&
		len(ignoreLabels) == 0 && len(selectLabels) == 0 &&
		!requireCI && !mustBeApproved {
		Logger.Warn("No filtering options specified. This will attempt to combine ALL open pull requests. Use  --labels, --ignore-labels, --branch-prefix, --branch-suffix, --branch-regex, --dependabot, etc to filter.")
	}

	return nil
}

func ValidateLabels(selectLabels []string, ignoreLabels []string) error {
	// Check for conflicts between ignoreLabels and selectLabels
	for _, ignoreL := range ignoreLabels {
		if i := slices.Index(selectLabels, ignoreL); i != -1 {
			return fmt.Errorf("%w: %q %q", errLabelsConflict, selectLabels[i], ignoreLabels)
		}
	}

	return nil
}
