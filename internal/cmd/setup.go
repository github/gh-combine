package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalContext creates a context that's cancelled on SIGINT or SIGTERM
func SetupSignalContext() (context.Context, context.CancelFunc) {
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
