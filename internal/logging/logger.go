package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"board-game-library/internal/config"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new structured logger based on configuration
func NewLogger(cfg config.LoggingConfig) (*Logger, error) {
	// Determine output writer
	var writer io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stdout", "":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		writer = file
	}

	// Determine log level
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text", "":
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	logger := slog.New(handler)
	return &Logger{Logger: logger}, nil
}

// WithComponent adds a component field to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{Logger: l.Logger.With("component", component)}
}

// WithRequestID adds a request ID field to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With("request_id", requestID)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With("error", err.Error())}
}

// LogStartup logs application startup information
func (l *Logger) LogStartup(cfg *config.Config) {
	l.Info("Starting Board Game Library",
		"server_port", cfg.Server.Port,
		"server_host", cfg.Server.Host,
		"database_path", cfg.Database.Path,
		"log_level", cfg.Logging.Level,
		"log_format", cfg.Logging.Format,
	)
}

// LogShutdown logs application shutdown information
func (l *Logger) LogShutdown() {
	l.Info("Board Game Library shutting down gracefully")
}

// LogDatabaseConnection logs database connection events
func (l *Logger) LogDatabaseConnection(dbPath string) {
	l.Info("Database connection established", "path", dbPath)
}

// LogDatabaseDisconnection logs database disconnection events
func (l *Logger) LogDatabaseDisconnection() {
	l.Info("Database connection closed")
}

// LogServerStart logs server start events
func (l *Logger) LogServerStart(addr string) {
	l.Info("HTTP server started", "address", addr)
}

// LogServerStop logs server stop events
func (l *Logger) LogServerStop() {
	l.Info("HTTP server stopped")
}