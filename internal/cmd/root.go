package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/github/gh-combine/internal/version"
)

func Run() error {
	Logger.Debug("starting gh-combine", "version", version.String())
	_, err := fmt.Println("Running gh-combine")
	if err != nil {
		return fmt.Errorf("oh no")
	}

	spinner := NewSpinner("")
	// Ensure the spinner is stopped before exiting the function.
	// If reference to `spinner` is changed, the reference to the new
	// spinner will be used to stop the spinner, so this works as
	// expected even in that case.
	defer func() { spinner.Stop() }()

	// Set up a channel to catch `Ctrl+C` (SIGINT) signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer signal.Stop(signalChan)

	// Start a goroutine to handle the signal
	go func() {
		<-signalChan
		Logger.Debug("Received SIGINT (Ctrl+C), stopping spinner and exiting...")
		spinner.Stop()
		os.Exit(1) // Exit with a non-zero code to indicate interruption
	}()

	// for initial testing, just sleep for 1 second and stop the spinner
	spinner.Suffix("running")
	time.Sleep(1 * time.Second)
	spinner.Stop()

	return nil
}
