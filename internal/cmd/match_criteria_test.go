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
