package logger

import (
	"log/slog"
	"os"
)

// This function creates a new logger with good defaults
func New() *slog.Logger {
	// Create options for the handler
	var options = &slog.HandlerOptions{
		// Info is a good default level for most applications
		Level: slog.LevelInfo,
	}

	// Create a text handler for the logger that outputs to stdout
	var handler = slog.NewTextHandler(os.Stdout, options)

	// Create a new logger with the handler
	var logger = slog.New(handler)

	// Return the logger
	return logger
}
