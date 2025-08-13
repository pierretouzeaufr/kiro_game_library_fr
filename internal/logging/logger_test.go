package logging

import (
	"testing"

	"board-game-library/internal/config"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config config.LoggingConfig
	}{
		{
			name: "text format to stdout",
			config: config.LoggingConfig{
				Level:  "info",
				Format: "text",
				Output: "stdout",
			},
		},
		{
			name: "json format to stderr",
			config: config.LoggingConfig{
				Level:  "debug",
				Format: "json",
				Output: "stderr",
			},
		},
		{
			name: "default values",
			config: config.LoggingConfig{
				Level:  "info",
				Format: "text",
				Output: "stdout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if err != nil {
				t.Fatalf("NewLogger() error = %v", err)
			}

			if logger == nil {
				t.Error("NewLogger() returned nil logger")
			}

			if logger.Logger == nil {
				t.Error("NewLogger() returned logger with nil slog.Logger")
			}
		})
	}
}

func TestLoggerWithComponent(t *testing.T) {
	config := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	componentLogger := logger.WithComponent("test-component")
	if componentLogger == nil {
		t.Error("WithComponent() returned nil logger")
	}

	if componentLogger.Logger == nil {
		t.Error("WithComponent() returned logger with nil slog.Logger")
	}
}

func TestLoggerWithRequestID(t *testing.T) {
	config := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	requestLogger := logger.WithRequestID("test-request-123")
	if requestLogger == nil {
		t.Error("WithRequestID() returned nil logger")
	}

	if requestLogger.Logger == nil {
		t.Error("WithRequestID() returned logger with nil slog.Logger")
	}
}

func TestLoggerWithError(t *testing.T) {
	config := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	testErr := &testError{message: "test error"}
	errorLogger := logger.WithError(testErr)
	if errorLogger == nil {
		t.Error("WithError() returned nil logger")
	}

	if errorLogger.Logger == nil {
		t.Error("WithError() returned logger with nil slog.Logger")
	}
}

func TestLogStartup(t *testing.T) {
	// Create a logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}

	logger, err := NewLogger(logConfig)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// Test configuration
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		Database: config.DatabaseConfig{
			Path: "./test.db",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}

	// This should not panic
	logger.LogStartup(testConfig)
}

func TestLogMethods(t *testing.T) {
	config := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// Test that these methods don't panic
	logger.LogShutdown()
	logger.LogDatabaseConnection("./test.db")
	logger.LogDatabaseDisconnection()
	logger.LogServerStart("localhost:8080")
	logger.LogServerStop()
}

func TestInvalidLogLevel(t *testing.T) {
	config := config.LoggingConfig{
		Level:  "invalid",
		Format: "text",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// Should default to info level and not panic
	if logger == nil {
		t.Error("NewLogger() returned nil logger for invalid level")
	}
}

func TestInvalidLogFormat(t *testing.T) {
	config := config.LoggingConfig{
		Level:  "info",
		Format: "invalid",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// Should default to text format and not panic
	if logger == nil {
		t.Error("NewLogger() returned nil logger for invalid format")
	}
}

// Helper type for testing error logging
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}