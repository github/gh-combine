//go:build !windows
// +build !windows

package tzlocal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const localZoneFile = "/etc/localtime" // symlinked file - set by OS

func inferFromPath(p string) (string, error) {
	var name string
	var err error

	parts := strings.Split(p, string(filepath.Separator))
	for i := range parts {
		if parts[i] == "zoneinfo" || parts[i] == "zoneinfo.default" {
			parts = parts[i+1:]
			break
		}
	}

	if len(parts) < 1 {
		err = fmt.Errorf("cannot infer timezone name from path: %q", p)
		return name, err
	}

	return filepath.Join(parts...), nil
}

// LocalTZ will run `/etc/localtime` and get the timezone from the resulting value `/usr/share/zoneinfo/America/New_York`
func LocalTZ() (string, error) {
	var name string
	fi, err := os.Lstat(localZoneFile)
	if err != nil {
		err = fmt.Errorf("failed to stat %q: %w", localZoneFile, err)
		return name, err
	}

	if (fi.Mode() & os.ModeSymlink) == 0 {
		err = fmt.Errorf("%q is not a symlink - cannot infer name", localZoneFile)
		return name, err
	}

	p, err := filepath.EvalSymlinks(localZoneFile)
	if err != nil {
		return name, err
	}

	// handles 1 & 2 part zone names
	name, err = inferFromPath(p)
	return name, err
}
