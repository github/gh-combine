//go:build windows && !go1.19
// +build windows,!go1.19

package tzlocal

import (
	"strings"

	"golang.org/x/sys/execabs"
)

// localTZfromTzutil executes command `tzutil /g` to get the name of the time zone Windows is configured to use.
func localTZfromTzutil() (string, error) {
	cmd := execabs.Command("tzutil", "/g")
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
