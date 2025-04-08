package cmd

import (
	"errors"
	"fmt"
	"slices"
)

// @grant: whenever possible, use sentinel errors. This ensures that those
// values can be testsed with `errors.Is()` correctly.
var (
	errNoRepo               = errors.New("must specify repositories or provide a file containing a list of repositories with --file")
	errLabelConflict        = errors.New("--ignore-label(s) and --label(s) cannot have the same value")
	errLabelsConflict       = errors.New("--label(s) contains a value which conflicts with --ignore-label(s)")
	errIgnoreLabelsConflict = errors.New("--ignore-label(s) contains a value which conflicts with --label(s)")
)

// validateInputs checks if the provided inputs are valid
func ValidateInputs(args []string) error {
	// @grant: using global variables is terrible for tests, so here I've
	// injected them into the function for validation.
	if err := ValidateLabels(selectLabel, selectLabels, ignoreLabel, ignoreLabels); err != nil {
		return err
	}

	// If no args and no file, we can't proceed
	if len(args) == 0 && reposFile == "" {
		return errNoRepo
	}

	// Warn if no filtering options are provided at all
	if branchPrefix == "" && branchSuffix == "" && branchRegex == "" &&
		ignoreLabel == "" && len(ignoreLabels) == 0 && selectLabel == "" && len(selectLabels) == 0 &&
		!requireCI && !mustBeApproved {
		Logger.Warn("No filtering options specified. This will attempt to combine ALL open pull requests.")
	}

	return nil
}

// @grant: this function is 100% self contained, and does the same logic as
// before, which makes it really easier to test.
func ValidateLabels(selectLabel string, selectLabels []string, ignoreLabel string, ignoreLabels []string) error {
	// Check that ignore-label and select-label are not the same
	if ignoreLabel != "" && selectLabel != "" && ignoreLabel == selectLabel {
		return errLabelConflict
	}

	// @grant: slices.Index beats a `for` and an `if`.
	// Check for conflicts between ignoreLabels and selectLabel
	if i := slices.Index(ignoreLabels, selectLabel); i != -1 {
		return fmt.Errorf("%w: %q %q", errIgnoreLabelsConflict, ignoreLabels[i], selectLabel)
	}

	// Check for conflicts between ignoreLabel and selectLabels
	if i := slices.Index(selectLabels, ignoreLabel); i != -1 {
		return fmt.Errorf("%w: %q %q", errLabelsConflict, selectLabels[i], ignoreLabel)
	}

	// @grant: no need to check for the length if we iterate directly, the
	// performance gain is moot.
	// Check for conflicts between ignoreLabels and selectLabels
	for _, ignoreL := range ignoreLabels {
		if i := slices.Index(selectLabels, ignoreL); i != -1 {
			return fmt.Errorf("%w: %q %q", errIgnoreLabelsConflict, ignoreLabels[i], selectLabel)
		}
	}

	return nil
}
