package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sachinthra/file-locker/backend/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates a new structured logger with log rotation
func New(cfg config.LoggingConfig) (*slog.Logger, error) {
	// Parse log level
	level := parseLevel(cfg.Level)

	// Setup log writer with rotation
	writer, err := setupWriter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup log writer: %w", err)
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Include source file and line number
	}

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(writer, opts)

	// Create logger
	logger := slog.New(handler)

	return logger, nil
}

// setupWriter configures the log writer with rotation using lumberjack
func setupWriter(cfg config.LoggingConfig) (io.Writer, error) {
	// Ensure log directory exists
	logDir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Configure lumberjack for log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSizeMB,  // megabytes
		MaxBackups: cfg.MaxBackups, // number of backups
		MaxAge:     cfg.MaxAgeDays, // days
		Compress:   true,           // compress rotated files
		LocalTime:  true,           // use local time for filenames
	}

	// Write to both file and stdout for better observability
	multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)

	return multiWriter, nil
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewCLILogger creates a logger for CLI with rotation at ~/.filelocker/cli.log
func NewCLILogger(level string) (*slog.Logger, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Setup CLI log path
	cliLogDir := filepath.Join(homeDir, ".filelocker")
	cliLogPath := filepath.Join(cliLogDir, "cli.log")

	// Ensure directory exists
	if err := os.MkdirAll(cliLogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create CLI log directory: %w", err)
	}

	// Configure lumberjack for CLI logs
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cliLogPath,
		MaxSize:    5,    // 5 MB
		MaxBackups: 3,    // keep 3 backups
		MaxAge:     14,   // 14 days
		Compress:   true, // compress old logs
		LocalTime:  true,
	}

	// CLI logs to file only (no stdout to avoid cluttering terminal)
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	handler := slog.NewJSONHandler(lumberjackLogger, opts)
	logger := slog.New(handler)

	return logger, nil
}
