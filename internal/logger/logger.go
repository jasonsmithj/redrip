// Package logger provides a centralized logging system for the redrip application.
// It uses Go's slog package to provide structured logging with configurable log levels.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

var (
	// Logger is the global logger instance
	Logger *slog.Logger
)

// Initialize sets up the global logger with the specified level
func Initialize(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var output io.Writer = os.Stdout

	// Create handler with specified level
	handler := slog.NewTextHandler(output, opts)
	Logger = slog.New(handler)

	slog.SetDefault(Logger)
}

// SetupLogger initializes the logger with INFO level
func SetupLogger() {
	Initialize(slog.LevelInfo)
}

// IsDebugMode returns true if the logger is configured for debug level
func IsDebugMode() bool {
	return Logger.Enabled(context.TODO(), slog.LevelDebug)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}
