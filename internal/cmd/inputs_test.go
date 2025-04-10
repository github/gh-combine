package cmd

import (
	"errors"
	"os"
	"testing"
)

func TestValidateLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		selectLabels []string
		ignoreLabels []string
		want         error
	}{
		{
			want: nil,
		},
		{
			selectLabels: []string{"a"},
			ignoreLabels: []string{"b"},
			want:         nil,
		},
		{
			selectLabels: []string{"a"},
			ignoreLabels: []string{"a", "b"},
			want:         errLabelsConflict,
		},
		{
			selectLabels: []string{"a", "b"},
			ignoreLabels: []string{"b"},
			want:         errLabelsConflict,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got := ValidateLabels(test.selectLabels, test.ignoreLabels)

			if !errors.Is(got, test.want) {
				t.Fatalf("want %s, but go %s", test.want, got)
			}
		})
	}
}

/*
// mockLogger creates a test logger that writes to a bytes.Buffer
func setupMockLogger() (*bytes.Buffer, func()) {
	var buf bytes.Buffer
	origLogger := Logger
	Logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return &buf, func() {
		Logger = origLogger
	}
}

func TestValidateInputs_NoReposSpecified(t *testing.T) {
	// Save original values
	origReposFile := reposFile
	defer func() { reposFile = origReposFile }()

	// Set up test values
	reposFile = ""

	// Run test
	err := ValidateInputs([]string{})

	// Verify results
	if err == nil {
		t.Errorf("Expected error when no repos specified, but got nil")
	}

	expectedErrMsg := "must specify repositories or provide a file"
	if err != nil && !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message containing %q, got: %q", expectedErrMsg, err.Error())
	}
}

func TestValidateInputs_SameIgnoreAndSelectLabel(t *testing.T) {
	// Save original values
	origIgnoreLabel := ignoreLabel
	origSelectLabel := selectLabel
	origReposFile := reposFile
	defer func() {
		ignoreLabel = origIgnoreLabel
		selectLabel = origSelectLabel
		reposFile = origReposFile
	}()

	// Set up test values
	ignoreLabel = "test-label"
	selectLabel = "test-label"
	reposFile = "dummy.txt" // To prevent the no-repos error

	// Run test
	err := ValidateInputs([]string{})

	// Verify results
	if err == nil {
		t.Errorf("Expected error when ignoreLabel and selectLabel have same value, but got nil")
	}

	expectedErrMsg := "--ignore-label(s) and --label(s) cannot have the same value"
	if err != nil && !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message containing %q, got: %q", expectedErrMsg, err.Error())
	}
}

func TestValidateInputs_NoFilteringOptions(t *testing.T) {
	// Save original values
	origBranchPrefix := branchPrefix
	origBranchSuffix := branchSuffix
	origBranchRegex := branchRegex
	origIgnoreLabel := ignoreLabel
	origSelectLabel := selectLabel
	origSelectLabels := selectLabels
	origRequireCI := requireCI
	origMustBeApproved := mustBeApproved
	origReposFile := reposFile
	defer func() {
		branchPrefix = origBranchPrefix
		branchSuffix = origBranchSuffix
		branchRegex = origBranchRegex
		ignoreLabel = origIgnoreLabel
		selectLabel = origSelectLabel
		selectLabels = origSelectLabels
		requireCI = origRequireCI
		mustBeApproved = origMustBeApproved
		reposFile = origReposFile
	}()

	// Set up test values - no filtering options
	branchPrefix = ""
	branchSuffix = ""
	branchRegex = ""
	ignoreLabel = ""
	selectLabel = ""
	selectLabels = nil
	requireCI = false
	mustBeApproved = false
	reposFile = "dummy.txt" // Prevent no-repos error

	// Set up mock logger
	logBuf, cleanup := setupMockLogger()
	defer cleanup()

	// Run test
	err := ValidateInputs([]string{"repo1"})

	// Verify results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "No filtering options specified") {
		t.Errorf("Expected warning about no filtering options, but it wasn't logged. Got: %s", logOutput)
	}
}

func TestValidateInputs_WithFilteringOptions(t *testing.T) {
	// Save original values and restore after test
	origBranchPrefix := branchPrefix
	origSelectLabels := selectLabels
	origReposFile := reposFile
	defer func() {
		branchPrefix = origBranchPrefix
		selectLabels = origSelectLabels
		reposFile = origReposFile
	}()

	// Set up test values - with filtering options
	branchPrefix = "feature/"
	selectLabels = []string{"enhancement"}
	reposFile = "dummy.txt"

	// Set up mock logger
	logBuf, cleanup := setupMockLogger()
	defer cleanup()

	// Run test
	err := ValidateInputs([]string{"owner/repo1"})

	// Verify results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	logOutput := logBuf.String()
	if strings.Contains(logOutput, "No filtering options specified") {
		t.Errorf("Warning about no filtering options was logged, but it shouldn't have been. Got: %s", logOutput)
	}
}

func TestValidateInputs_WithIgnoreLabels(t *testing.T) {
	// Save original values
	origIgnoreLabels := ignoreLabels
	origSelectLabel := selectLabel
	origReposFile := reposFile
	defer func() {
		ignoreLabels = origIgnoreLabels
		selectLabel = origSelectLabel
		reposFile = origReposFile
	}()

	// Test case where ignoreLabels contains the same value as selectLabel
	ignoreLabels = []string{"bug", "enhancement"}
	selectLabel = "enhancement"
	reposFile = "dummy.txt"

	err := ValidateInputs([]string{"repo1"})

	if err == nil {
		t.Errorf("Expected error when ignoreLabels contains a value from selectLabel, but got nil")
	}
}

func TestValidateInputs_WithSelectLabelsAndIgnoreLabels(t *testing.T) {
	// Save original values
	origIgnoreLabels := ignoreLabels
	origSelectLabels := selectLabels
	origReposFile := reposFile
	defer func() {
		ignoreLabels = origIgnoreLabels
		selectLabels = origSelectLabels
		reposFile = origReposFile
	}()

	// Test case where ignoreLabels and selectLabels have common values
	ignoreLabels = []string{"bug", "enhancement"}
	selectLabels = []string{"documentation", "enhancement"}
	reposFile = "dummy.txt"

	err := ValidateInputs([]string{"repo1"})

	if err == nil {
		t.Errorf("Expected error when ignoreLabels and selectLabels have common values, but got nil")
	}
}

func TestValidateInputs_ValidInputs(t *testing.T) {
	// Test with valid inputs
	tests := []struct {
		name           string
		args           []string
		reposFile      string
		branchPrefix   string
		ignoreLabel    string
		selectLabel    string
		ignoreLabels   []string
		selectLabels   []string
		requireCI      bool
		mustBeApproved bool
	}{
		{
			name:         "With args",
			args:         []string{"owner/repo1", "owner/repo2"},
			branchPrefix: "feature/",
		},
		{
			name:        "With file",
			args:        []string{},
			reposFile:   "repos.txt",
			selectLabel: "enhancement",
		},
		{
			name:           "With multiple filtering options",
			args:           []string{"owner/repo"},
			branchPrefix:   "feature/",
			requireCI:      true,
			mustBeApproved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origReposFile := reposFile
			origBranchPrefix := branchPrefix
			origIgnoreLabel := ignoreLabel
			origSelectLabel := selectLabel
			origIgnoreLabels := ignoreLabels
			origSelectLabels := selectLabels
			origRequireCI := requireCI
			origMustBeApproved := mustBeApproved
			defer func() {
				reposFile = origReposFile
				branchPrefix = origBranchPrefix
				ignoreLabel = origIgnoreLabel
				selectLabel = origSelectLabel
				ignoreLabels = origIgnoreLabels
				selectLabels = origSelectLabels
				requireCI = origRequireCI
				mustBeApproved = origMustBeApproved
			}()

			// Set test values
			reposFile = tt.reposFile
			branchPrefix = tt.branchPrefix
			ignoreLabel = tt.ignoreLabel
			selectLabel = tt.selectLabel
			ignoreLabels = tt.ignoreLabels
			selectLabels = tt.selectLabels
			requireCI = tt.requireCI
			mustBeApproved = tt.mustBeApproved

			// Run test
			err := ValidateInputs(tt.args)

			// Verify results
			if err != nil {
				t.Errorf("Expected no error with valid inputs, got: %v", err)
			}
		})
	}
}
*/

// TestMain sets up and tears down the testing environment
func TestMain(m *testing.M) {
	// Save original logger
	origLogger := Logger

	// Run tests
	code := m.Run()

	// Restore original logger
	Logger = origLogger

	os.Exit(code)
}
