package service

import (
	"log/slog"
	"os"
)

// newTestLogger creates a logger for tests that discards output or logs to stderr.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}
