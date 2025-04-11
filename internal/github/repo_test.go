package github

import (
	"errors"
	"testing"
)

func TestParseRepo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		repo string
		err  error
		want Repo
	}{
		{
			err: ErrInvalidRepository,
		},
		{
			repo: "owner",
			err:  ErrInvalidRepository,
		},
		{
			repo: "/repo",
			err:  ErrInvalidRepository,
		},
		{
			repo: "/",
			err:  ErrInvalidRepository,
		},

		{
			repo: "owner/repo",
			want: Repo{Owner: "owner", Repo: "repo"},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got, err := ParseRepo(test.repo)
			if !errors.Is(err, test.err) {
				t.Errorf("want %q, got %q", test.err, err)
			}

			if got.Owner != test.want.Owner {
				t.Errorf("want owner %s, got %s", test.want.Owner, got.Owner)
			}

			if got.Repo != test.want.Repo {
				t.Errorf("want owner %s, got %s", test.want.Repo, got.Repo)
			}
		})
	}
}
