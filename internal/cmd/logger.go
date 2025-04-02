package cmd

import (
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger

func init() {
	// Set log level based on environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // default to info level
	}

	// Upcase the log level
	logLevel = strings.ToUpper(logLevel)

	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   false,              // Include source file and line number for context
		ReplaceAttr: humanFriendlyAttrs, // Customize log attributes for readability
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	// Configure slog with the specified log level
	Logger = slog.New(handler)
}

// humanFriendlyAttrs customizes log attributes for human readability
func humanFriendlyAttrs(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.LevelKey:
		// Capitalize log levels for better visibility
		a.Value = slog.StringValue(strings.ToUpper(a.Value.String()))
	}
	return a
}
