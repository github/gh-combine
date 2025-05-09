package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/github/gh-combine/internal/github"
)

func reposEqual(a, b []github.Repo) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Owner != b[i].Owner || a[i].Repo != b[i].Repo {
			return false
		}
	}

	return true
}

func TestParseRepositoriesArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		args []string
		want []github.Repo
		err  error
	}{
		{},

		{
			args: []string{"a/b"},
			want: []github.Repo{
				{Owner: "a", Repo: "b"},
			},
		},

		{
			args: []string{"a/b", "c/d"},
			want: []github.Repo{
				{Owner: "a", Repo: "b"},
				{Owner: "c", Repo: "d"},
			},
		},

		{
			args: []string{"a"},
			err:  github.ErrInvalidRepository,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got, err := parseRepositoriesArgs(test.args)

			if !errors.Is(err, test.err) {
				t.Errorf("want error %q, got %q", test.err, err)
			}

			if !reposEqual(test.want, got) {
				t.Errorf("want %q, got %q", test.want, got)
			}
		})
	}
}

func TestParseRepositoriesFile(t *testing.T) {
	tests := []struct {
		content string
		want    []github.Repo
		err     error
	}{
		{},

		{
			content: `
# Comment line
owner1/repo1
   # Indented comment
   owner2/repo2 # comment after repo
# Another comment

    # Indented comment after empty line
`,
			want: []github.Repo{
				{Owner: "owner1", Repo: "repo1"},
				{Owner: "owner2", Repo: "repo2"},
			},
		},

		{
			content: `
owner1 # invalid repo
`,
			err: github.ErrInvalidRepository,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			fp := filepath.Join(t.TempDir(), "repos")
			if err := os.WriteFile(fp, []byte(test.content), 0o644); err != nil {
				t.Fatalf("failed to write file %s: %v", fp, err)
			}

			got, err := parseRepositoriesFile(fp)

			if !errors.Is(err, test.err) {
				t.Errorf("want error %q, got %q", test.err, err)
			}

			if !reposEqual(test.want, got) {
				t.Errorf("want %q, got %q", test.want, got)
			}
		})
	}
}
