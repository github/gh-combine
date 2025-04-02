package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestApplyDefaultOwner(t *testing.T) {
	tests := []struct {
		name         string
		repo         string
		defaultOwner string
		want         string
	}{
		{
			name:         "empty default owner",
			repo:         "repo1",
			defaultOwner: "",
			want:         "repo1",
		},
		{
			name:         "repo already has owner",
			repo:         "octocat/repo1",
			defaultOwner: "another-owner",
			want:         "octocat/repo1",
		},
		{
			name:         "repo needs default owner",
			repo:         "repo1",
			defaultOwner: "octocat",
			want:         "octocat/repo1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyDefaultOwner(tt.repo, tt.defaultOwner)
			if got != tt.want {
				t.Errorf("applyDefaultOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRepositories(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	validFilePath := filepath.Join(tempDir, "valid.txt")
	validFileContent := `owner1/repo1
owner2/repo2
# This is a comment
repo3
   
  repo4  
`
	err := os.WriteFile(validFilePath, []byte(validFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		reposFile    string
		defaultOwner string
		want         []string
		wantErr      bool
	}{
		{
			name:         "empty inputs",
			args:         []string{},
			reposFile:    "",
			defaultOwner: "",
			want:         []string{},
			wantErr:      false,
		},
		{
			name:         "single repository arg",
			args:         []string{"owner/repo"},
			reposFile:    "",
			defaultOwner: "",
			want:         []string{"owner/repo"},
			wantErr:      false,
		},
		{
			name:         "multiple repository args",
			args:         []string{"owner/repo1", "owner/repo2"},
			reposFile:    "",
			defaultOwner: "",
			want:         []string{"owner/repo1", "owner/repo2"},
			wantErr:      false,
		},
		{
			name:         "comma-separated repository args",
			args:         []string{"owner/repo1,owner/repo2", "owner/repo3"},
			reposFile:    "",
			defaultOwner: "",
			want:         []string{"owner/repo1", "owner/repo2", "owner/repo3"},
			wantErr:      false,
		},
		{
			name:         "with default owner",
			args:         []string{"repo1", "owner2/repo2"},
			reposFile:    "",
			defaultOwner: "owner1",
			want:         []string{"owner1/repo1", "owner2/repo2"},
			wantErr:      false,
		},
		{
			name:         "with valid file",
			args:         []string{},
			reposFile:    validFilePath,
			defaultOwner: "default-owner",
			want:         []string{"owner1/repo1", "owner2/repo2", "default-owner/repo3", "default-owner/repo4"},
			wantErr:      false,
		},
		{
			name:         "with invalid file",
			args:         []string{},
			reposFile:    filepath.Join(tempDir, "nonexistent.txt"),
			defaultOwner: "",
			want:         nil,
			wantErr:      true,
		},
		{
			name:         "with both args and file",
			args:         []string{"arg-repo1", "owner/arg-repo2"},
			reposFile:    validFilePath,
			defaultOwner: "default-owner",
			want:         []string{"default-owner/arg-repo1", "owner/arg-repo2", "owner1/repo1", "owner2/repo2", "default-owner/repo3", "default-owner/repo4"},
			wantErr:      false,
		},
		{
			name:         "with whitespace in comma-separated repos",
			args:         []string{"repo1, repo2,  repo3"},
			reposFile:    "",
			defaultOwner: "owner",
			want:         []string{"owner/repo1", "owner/repo2", "owner/repo3"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepositories(tt.args, tt.reposFile, tt.defaultOwner)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRepositories() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestInvalidFileFormat ensures that whitespace and comments are handled correctly
func TestParseRepositoriesFileFormat(t *testing.T) {
	// Create a test file with various edge cases
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "format_test.txt")
	fileContent := `
# Comment line
owner1/repo1
   # Indented comment
   owner2/repo2   
# Another comment

repo3  # Comment after repo
  
    # Indented comment with spaces
    
`
	err := os.WriteFile(testFilePath, []byte(fileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	want := []string{"owner1/repo1", "owner2/repo2", "default/repo3"}
	got, err := ParseRepositories([]string{}, testFilePath, "default")
	if err != nil {
		t.Fatalf("ParseRepositories() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseRepositories() with complex file format = %v, want %v", got, want)
	}
}

// Additional test for empty entries in comma-separated list
func TestParseRepositoriesWithEmptyEntries(t *testing.T) {
	args := []string{"repo1,,repo2,", ",repo3"}
	defaultOwner := "owner"

	want := []string{"owner/repo1", "owner/repo2", "owner/repo3"}
	got, err := ParseRepositories(args, "", defaultOwner)
	if err != nil {
		t.Fatalf("ParseRepositories() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseRepositories() with empty entries = %v, want %v", got, want)
	}
}
