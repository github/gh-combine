package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/github/gh-combine/internal/version"
)

// Run executes the main functionality of the application.
func Run() error {
	ctx, cancel := setupSignalContext()
	defer cancel()

	Logger.Debug("starting gh-combine", "version", version.String())

	spinner := NewSpinner("")
	defer spinner.Stop()

	if err := executeCommand(ctx, spinner); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// setupSignalContext creates a context that's cancelled on SIGINT or SIGTERM.
func setupSignalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-signalChan:
			Logger.Debug("Received interrupt signal, cancelling operations...")
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(signalChan)
	}()

	return ctx, cancel
}

// executeCommand performs the actual API call and processing.
func executeCommand(ctx context.Context, spinner *Spinner) error {
	// Create REST client
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Define response structure
	var response []struct {
		Name string
	}

	// Make the API request
	if err := restClient.Get("repos/cli/cli/tags", &response); err != nil {
		return fmt.Errorf("REST API request failed: %w", err)
	}

	Logger.Debug("response", "response", response)
	return nil
}
