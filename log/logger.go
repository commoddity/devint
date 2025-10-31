package log

import (
	"log/slog"
	"os"
)

// NewJSONLogger creates a new slog logger that outputs JSON format.
func NewJSONLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}
