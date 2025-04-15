package common

import (
	"reflect"
	"testing"
)

func TestNormalizeArray(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "Empty array",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "Already lowercase",
			input: []string{"test", "already", "lowercase"},
			want:  []string{"test", "already", "lowercase"},
		},
		{
			name:  "Mixed case",
			input: []string{"Test", "UPPERCASE", "lowercase", "MixedCase"},
			want:  []string{"test", "uppercase", "lowercase", "mixedcase"},
		},
		{
			name:  "Special characters",
			input: []string{"TEST-123", "Feature/branch", "BUG_FIX"},
			want:  []string{"test-123", "feature/branch", "bug_fix"},
		},
		{
			name:  "Single item",
			input: []string{"SINGLE"},
			want:  []string{"single"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeArray(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NormalizeArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
