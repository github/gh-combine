package cmd

import "testing"

func TestLabelsMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		prLabels     []string
		ignoreLabels []string
		selectLabels []string
		want         bool
	}{
		{
			want: true,
		},

		{
			prLabels:     []string{"a", "b"},
			ignoreLabels: []string{"b"},
			want:         false,
		},
		{
			prLabels:     []string{"a", "b"},
			ignoreLabels: []string{"b", "c"},
			want:         false,
		},

		{
			prLabels:     []string{"a"},
			ignoreLabels: []string{"b"},
			selectLabels: []string{"c"},
			want:         false,
		},
		{
			prLabels:     []string{"a", "c"},
			ignoreLabels: []string{"b"},
			selectLabels: []string{"c"},
			want:         true,
		},
		{
			prLabels:     []string{"a"},
			ignoreLabels: []string{"b"},
			selectLabels: []string{"a", "c"},
			want:         true,
		},
		{
			prLabels:     []string{},
			ignoreLabels: []string{"b"},
			selectLabels: []string{"a", "c"},
			want:         false,
		},
		{
			prLabels:     []string{},
			ignoreLabels: []string{"b"},
			selectLabels: []string{},
			want:         true,
		},
		{
			prLabels:     []string{},
			ignoreLabels: []string{},
			selectLabels: []string{"a"},
			want:         false,
		},
		{
			prLabels:     []string{},
			ignoreLabels: []string{},
			selectLabels: []string{},
			want:         true,
		},
		{
			prLabels:     []string{"a"},
			selectLabels: []string{"a"},
			ignoreLabels: []string{"b"},
			want:         true,
		},
		{
			prLabels:     []string{"a"},
			selectLabels: []string{"a"},
			ignoreLabels: []string{"a"},
			want:         false,
		},
		{
			prLabels:     []string{"a"},
			selectLabels: []string{},
			ignoreLabels: []string{},
			want:         true,
		},
		{
			prLabels:     []string{"a"},
			selectLabels: []string{"a"},
			ignoreLabels: []string{},
			want:         true,
		},
		{
			prLabels:     []string{"a", "b", "c"},
			selectLabels: []string{"a", "b"},
			ignoreLabels: []string{"c"},
			want:         false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got := labelsMatch(test.prLabels, test.ignoreLabels, test.selectLabels)
			if got != test.want {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}
func TestBranchMatchesCriteria(t *testing.T) {
	t.Parallel()

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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Set global variables for the test
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
