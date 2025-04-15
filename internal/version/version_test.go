package version

import (
	"runtime/debug"
	"testing"
)

func TestString(t *testing.T) {
	// Save original values
	origTag := tag
	origCommit := commit
	origDate := date
	origBuildInfoReader := buildInfoReader
	
	// Restore original values after the test
	defer func() {
		tag = origTag
		commit = origCommit
		date = origDate
		buildInfoReader = origBuildInfoReader
	}()
	
	// Test 1: With preset values (simulating ldflags setting)
	t.Run("with preset values", func(t *testing.T) {
		// Set known values for testing
		tag = "v1.0.0"
		commit = "abc123"
		date = "2025-04-15"
		
		// Mock the buildInfoReader to return false so that preset values are used
		buildInfoReader = func() (*debug.BuildInfo, bool) {
			return nil, false
		}
		
		result := String()
		
		// Test the full format
		expected := "v1.0.0 (abc123) built at 2025-04-15\nhttps://github.com/github/gh-combine/releases/tag/v1.0.0"
		if result != expected {
			t.Errorf("Expected version string to be:\n%q\nbut got:\n%q", expected, result)
		}
	})
	
	// Test 2: With mock build info that updates commit and date
	t.Run("with mock build info", func(t *testing.T) {
		// Set initial values
		tag = "dev"
		commit = "initial-commit"
		date = "initial-date"
		
		// Create mock build info with specific values
		mockSettings := []debug.BuildSetting{
			{Key: "vcs.revision", Value: "mock-commit-hash"},
			{Key: "vcs.time", Value: "mock-build-time"},
			{Key: "other.key", Value: "other-value"},
		}
		
		buildInfoReader = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Settings: mockSettings,
			}, true
		}
		
		result := String()
		
		// Check if the values from build info were used
		expected := "dev (mock-commit-hash) built at mock-build-time\nhttps://github.com/github/gh-combine/releases/tag/dev"
		if result != expected {
			t.Errorf("Expected version string to be:\n%q\nbut got:\n%q", expected, result)
		}
	})
	
	// Test 3: With empty build info settings
	t.Run("with empty build info settings", func(t *testing.T) {
		// Set initial values
		tag = "dev"
		commit = "unchanged-commit"
		date = "unchanged-date"
		
		// Empty build settings
		buildInfoReader = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Settings: []debug.BuildSetting{},
			}, true
		}
		
		result := String()
		
		// The values should remain unchanged
		expected := "dev (unchanged-commit) built at unchanged-date\nhttps://github.com/github/gh-combine/releases/tag/dev"
		if result != expected {
			t.Errorf("Expected version string to be:\n%q\nbut got:\n%q", expected, result)
		}
	})
}