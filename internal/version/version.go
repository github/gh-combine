package version

import (
	"fmt"
	"runtime/debug"
)

var (
	tag    = "dev" // set via ldflags
	commit = "123abc"
	date   = "now"
)

const template = "%s (%s) built at %s\nhttps://github.com/github/gh-combine/releases/tag/%s"

// buildInfoReader is a function type that can be mocked in tests
var buildInfoReader = defaultBuildInfoReader

// defaultBuildInfoReader is the actual implementation using debug.ReadBuildInfo
func defaultBuildInfoReader() (*debug.BuildInfo, bool) {
	return debug.ReadBuildInfo()
}

func String() string {
	info, ok := buildInfoReader()

	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value
			}
			if setting.Key == "vcs.time" {
				date = setting.Value
			}
		}
	}

	return fmt.Sprintf(template, tag, commit, date, tag)
}
