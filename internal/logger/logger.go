// Package logger provides structured logging using Go's slog package.
package logger

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	// Create a structured logger with JSON output
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// SetLevel sets the logging level
func SetLevel(level slog.Level) {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}

// With returns a logger with the given attributes
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}
