package logger

import (
	"io"
	"log/slog"
)

// SetupTestLogger initializes a test logger that writes to the provided writer
// or to io.Discard if nil is provided
func SetupTestLogger(w io.Writer) {
	if w == nil {
		w = io.Discard
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewTextHandler(w, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// InitNullLogger sets up a logger that discards all output
// Useful for tests to avoid log spam
func InitNullLogger() {
	SetupTestLogger(nil)
}
