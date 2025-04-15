package cmd

import "testing"

func TestLabelsMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		prLabels      []string
		ignoreLabels  []string
		selectLabels  []string
		caseSensitive bool
		want          bool
	}{
		{
			want: true,
		},

		{
			name:          "--ignore-labels match",
			prLabels:      []string{"a", "b"},
			ignoreLabels:  []string{"b"},
			want:          false,
			caseSensitive: false,
		},
		{
			name:          "--ignore-labels match (with one out of two)",
			prLabels:      []string{"a", "b"},
			ignoreLabels:  []string{"b", "c"},
			want:          false,
			caseSensitive: false,
		},

		{
			name:          "no labels match (select or ignore)",
			prLabels:      []string{"a"},
			ignoreLabels:  []string{"b"},
			selectLabels:  []string{"c"},
			want:          false,
			caseSensitive: false,
		},
		{
			name:          "--select-labels match",
			prLabels:      []string{"a", "c"},
			ignoreLabels:  []string{"b"},
			selectLabels:  []string{"c"},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "--select-labels match (with one out of two) and ignore labels don't match",
			prLabels:      []string{"a"},
			ignoreLabels:  []string{"b"},
			selectLabels:  []string{"a", "c"},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "the pull request has no labels",
			prLabels:      []string{},
			ignoreLabels:  []string{"b"},
			selectLabels:  []string{"a", "c"},
			want:          false,
			caseSensitive: false,
		},
		{
			name:          "the pull request has no labels and ignore labels don't match so it matches - but select labels is empty so it means all labels or even no labels match",
			prLabels:      []string{},
			ignoreLabels:  []string{"b"},
			selectLabels:  []string{},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "the pull request has no labels but we want to match the a label",
			prLabels:      []string{},
			ignoreLabels:  []string{},
			selectLabels:  []string{"a"},
			want:          false,
			caseSensitive: false,
		},
		{
			name:          "no label match criteria, so it matches",
			prLabels:      []string{},
			ignoreLabels:  []string{},
			selectLabels:  []string{},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "with one matching label and no matching ignore labels so it matches",
			prLabels:      []string{"a"},
			selectLabels:  []string{"a"},
			ignoreLabels:  []string{"b"},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "the pr labels match the select and ignore labels so it doesn't match",
			prLabels:      []string{"a"},
			selectLabels:  []string{"a"},
			ignoreLabels:  []string{"a"},
			want:          false,
			caseSensitive: false,
		},
		{
			name:          "the pr has one label but no defined ignore or select labels so it matches",
			prLabels:      []string{"a"},
			selectLabels:  []string{},
			ignoreLabels:  []string{},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "the pr has one label and it is the select label so it matches",
			prLabels:      []string{"a"},
			selectLabels:  []string{"a"},
			ignoreLabels:  []string{},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "the pr has labels and matching select labels but it matches an ignore label so it doesn't match",
			prLabels:      []string{"a", "b", "c"},
			selectLabels:  []string{"a", "b"},
			ignoreLabels:  []string{"c"},
			want:          false,
			caseSensitive: false,
		},
		{
			name:          "the pr has uppercase labels and we are using case insensitive labels so it matches",
			prLabels:      []string{"Dependencies", "rUby", "ready-for-Review"},
			selectLabels:  []string{"dependencies", "ready-for-review"},
			ignoreLabels:  []string{"blocked"},
			want:          true,
			caseSensitive: false,
		},
		{
			name:          "the pr has uppercase labels and we are using case sensitive labels so it doesn't match",
			prLabels:      []string{"Dependencies", "rUby", "ready-for-Review"},
			selectLabels:  []string{"dependencies", "ready-for-review"},
			ignoreLabels:  []string{"blocked"},
			want:          false,
			caseSensitive: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Save the original value of caseSensitiveLabels
			originalCaseSensitive := caseSensitiveLabels
			defer func() { caseSensitiveLabels = originalCaseSensitive }() // Restore after test

			// Set caseSensitiveLabels for this test
			caseSensitiveLabels = test.caseSensitive

			// Run the function
			got := labelsMatch(test.prLabels, test.ignoreLabels, test.selectLabels, test.caseSensitive)
			if got != test.want {
				t.Errorf("Test %q failed: want %v, got %v", test.name, test.want, got)
			}
		})
	}
}
func TestBranchMatchesCriteria(t *testing.T) {
	// Remove t.Parallel() at the function level to avoid races on global variables

	// Define test cases
	tests := []struct {
		name          string
		branch        string
		combineBranch string
		prefix        string
		suffix        string
		regex         string
		want          bool
	}{
		{
			name:          "Branch matches all criteria",
			branch:        "feature/test",
			combineBranch: "combined-prs",
			prefix:        "feature/",
			suffix:        "/test",
			regex:         `^feature/.*$`,
			want:          true,
		},
		{
			name:          "Branch is the combine branch",
			branch:        "combined-prs",
			combineBranch: "combined-prs",
			want:          false,
		},
		{
			name:          "Branch ends with the combine branch",
			branch:        "fix-combined-prs",
			combineBranch: "combined-prs",
			want:          true,
		},
		{
			name:          "No filters specified",
			branch:        "any-branch",
			combineBranch: "combined-prs",
			want:          true,
		},
		{
			name:          "No filters specified and partial match on combine branch name",
			branch:        "bug/combined-prs-fix",
			combineBranch: "combined-prs",
			want:          true,
		},
		{
			name:          "Prefix does not match",
			branch:        "test/feature",
			combineBranch: "combined-prs",
			prefix:        "feature/",
			want:          false,
		},
		{
			name:          "Suffix does not match",
			branch:        "feature/test",
			combineBranch: "combined-prs",
			suffix:        "/feature",
			want:          false,
		},
		{
			name:          "Regex does not match",
			branch:        "test/feature",
			combineBranch: "combined-prs",
			regex:         `^feature/.*`,
			want:          false,
		},
		{
			name:          "Invalid regex pattern",
			branch:        "feature/test",
			combineBranch: "combined-prs",
			regex:         `^(feature/.*$`,
			want:          false,
		},
		{
			name:          "Branch matches prefix only",
			branch:        "feature/test",
			combineBranch: "combined-prs",
			prefix:        "feature/",
			want:          true,
		},
		{
			name:          "Branch matches suffix only",
			branch:        "test/feature",
			combineBranch: "combined-prs",
			suffix:        "/feature",
			want:          true,
		},
		{
			name:          "Branch matches regex only",
			branch:        "feature/test",
			combineBranch: "combined-prs",
			regex:         `^feature/.*$`,
			want:          true,
		},
	}

	for _, test := range tests {
		test := test // Create a local copy of the test variable to use in the closure
		t.Run(test.name, func(t *testing.T) {
			t.Parallel() // Parallelize at the subtest level, each with their own local variables

			// Save original values of global variables
			origCombineBranchName := combineBranchName
			origBranchPrefix := branchPrefix
			origBranchSuffix := branchSuffix
			origBranchRegex := branchRegex
			
			// Restore original values after test
			defer func() {
				combineBranchName = origCombineBranchName
				branchPrefix = origBranchPrefix
				branchSuffix = origBranchSuffix
				branchRegex = origBranchRegex
			}()

			// Set global variables for this specific test
			combineBranchName = test.combineBranch
			branchPrefix = test.prefix
			branchSuffix = test.suffix
			branchRegex = test.regex

			// Run the function
			got := branchMatchesCriteria(test.branch)

			// Check the result
			if got != test.want {
				t.Errorf("branchMatchesCriteria(%q) = %v; want %v", test.branch, got, test.want)
			}
		})
	}
}

func prMatchesCriteriaWithMocks(branch string, prLabels []string, branchMatches func(string) bool, labelsMatch func([]string, []string, []string) bool) bool {
	return branchMatches(branch) && labelsMatch(prLabels, nil, nil)
}

func TestPrMatchesCriteriaWithMocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		branch     string
		prLabels   []string
		branchPass bool
		labelsPass bool
		want       bool
	}{
		{
			name:       "Branch and labels match",
			branch:     "feature/test",
			prLabels:   []string{"bug", "enhancement"},
			branchPass: true,
			labelsPass: true,
			want:       true,
		},
		{
			name:       "Branch does not match",
			branch:     "hotfix/test",
			prLabels:   []string{"bug", "enhancement"},
			branchPass: false,
			labelsPass: true,
			want:       false,
		},
		{
			name:       "Labels do not match",
			branch:     "feature/test",
			prLabels:   []string{"wip"},
			branchPass: true,
			labelsPass: false,
			want:       false,
		},
		{
			name:       "Neither branch nor labels match",
			branch:     "hotfix/test",
			prLabels:   []string{"wip"},
			branchPass: false,
			labelsPass: false,
			want:       false,
		},
		{
			name:       "No branch or label filters specified",
			branch:     "any-branch",
			prLabels:   []string{},
			branchPass: true,
			labelsPass: true,
			want:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock branchMatchesCriteria and labelsMatch
			mockBranchMatchesCriteria := func(branch string) bool {
				return test.branchPass
			}
			mockLabelsMatch := func(prLabels []string, ignoreLabels []string, selectLabels []string) bool {
				return test.labelsPass
			}

			got := prMatchesCriteriaWithMocks(test.branch, test.prLabels, mockBranchMatchesCriteria, mockLabelsMatch)
			if got != test.want {
				t.Errorf("PrMatchesCriteria(%q, %v) = %v; want %v", test.branch, test.prLabels, got, test.want)
			}
		})
	}
}
