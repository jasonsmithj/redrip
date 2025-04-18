package logger

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func init() {
	// Initialize a null logger by default for all tests
	InitNullLogger()
}

func TestInitialize(t *testing.T) {
	// Redirect logs to a buffer for testing
	var buf bytes.Buffer
	oldLogger := Logger
	defer func() { Logger = oldLogger }()

	// Initialize with Debug level
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	// Log messages at different levels
	Debug("debug message", "key", "value")
	Info("info message", "key", "value")
	Warn("warn message", "key", "value")
	Error("error message", "key", "value")

	// Check that all messages were logged
	logOutput := buf.String()
	if !strings.Contains(logOutput, "debug message") {
		t.Error("Debug message not found in logs")
	}
	if !strings.Contains(logOutput, "info message") {
		t.Error("Info message not found in logs")
	}
	if !strings.Contains(logOutput, "warn message") {
		t.Error("Warn message not found in logs")
	}
	if !strings.Contains(logOutput, "error message") {
		t.Error("Error message not found in logs")
	}

	// Reset the buffer
	buf.Reset()

	// Initialize with Info level
	handler = slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Logger = slog.New(handler)

	// Log messages at different levels
	Debug("debug message", "key", "value")
	Info("info message", "key", "value")
	Warn("warn message", "key", "value")
	Error("error message", "key", "value")

	// Check that debug message was not logged
	logOutput = buf.String()
	if strings.Contains(logOutput, "debug message") {
		t.Error("Debug message found in logs when level is Info")
	}
	if !strings.Contains(logOutput, "info message") {
		t.Error("Info message not found in logs")
	}
}

func TestSetupLogger(t *testing.T) {
	// Save original stdout and redirect logs to a buffer for testing
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore original stdout at the end of the test
	defer func() { os.Stdout = origStdout }()

	// Call the function we're testing
	SetupLogger()

	// Log a message
	Info("test message")

	// Close the writer and read the output
	err := w.Close()
	if err != nil {
		t.Errorf("Failed to close writer: %v", err)
		return
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	logOutput := buf.String()

	// Check that the message was logged with the correct level and format
	if !strings.Contains(logOutput, "INFO") && !strings.Contains(logOutput, "test message") {
		t.Error("Expected log message not found")
	}
}

func TestSetupTestLogger(t *testing.T) {
	// Test with custom writer
	var buf bytes.Buffer

	// Save original logger
	oldLogger := Logger
	defer func() { Logger = oldLogger }()

	// Setup test logger with our buffer
	SetupTestLogger(&buf)

	// Log a message
	Info("test custom writer")

	// Check the message was written to our buffer
	if !strings.Contains(buf.String(), "test custom writer") {
		t.Error("Message not written to custom writer")
	}

	// Test null logger
	InitNullLogger()

	// Reset the buffer
	buf.Reset()

	// Write to the discarded logger
	Info("this should not appear")

	// Check nothing was written to our buffer
	if buf.Len() > 0 {
		t.Error("Message was written when using null logger")
	}
}
