package cmd

import (
	"fmt"

	"github.com/github/gh-combine/internal/version"
)

func Run() error {
	Logger.Debug("starting gh-combine", "version", version.String())
	_, err := fmt.Println("Running gh-combine")
	if err != nil {
		return fmt.Errorf("oh no")
	}

	return nil
}
