package cmd

import "errors"

// validateInputs checks if the provided inputs are valid
func ValidateInputs(args []string) error {
	// Check that ignore-label and select-label are not the same
	if ignoreLabel != "" && selectLabel != "" && ignoreLabel == selectLabel {
		return errors.New("--ignore-label(s) and --label(s) cannot have the same value")
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
