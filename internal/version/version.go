package version

import (
	"fmt"
	"runtime/debug"
)

var (
	tag       = "dev" // set via ldflags
	commit    = "123abc"
	buildTime = "now"
)

const template = "%s (%s) built at %s\nhttps://github.com/github/gh-combine/releases/tag/%s"

// buildInfoReader is a function type that can be mocked in tests
var buildInfoReader = defaultBuildInfoReader

// defaultBuildInfoReader is the actual implementation using debug.ReadBuildInfo
func defaultBuildInfoReader() (*debug.BuildInfo, bool) {
	return debug.ReadBuildInfo()
}

func String() string {
	// Start with ldflags values
	currentCommit := commit
	currentDate := buildTime

	// Override with VCS info if available and ldflags weren't set
	info, ok := buildInfoReader()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && commit == "123abc" {
				currentCommit = setting.Value
			}
			if setting.Key == "vcs.time" && buildTime == "now" {
				currentDate = setting.Value
			}
		}
	}

	return fmt.Sprintf(template, tag, currentCommit, currentDate, tag)
}
