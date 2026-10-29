// Package logger configures the process-wide slog logger.
package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// Init builds a slog logger that writes to stdout and installs it as the default one.
// level is one of debug, info, warn, error (case-insensitive); asJSON picks the JSON
// handler over the human-readable text one.
func Init(level string, asJSON bool) error {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if asJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))

	return nil
}
