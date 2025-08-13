package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test with default values
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify default values
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host '0.0.0.0', got %s", config.Server.Host)
	}

	if config.Database.Path != "./data/library.db" {
		t.Errorf("Expected default database path './data/library.db', got %s", config.Database.Path)
	}

	if config.Alerts.ReminderDays != 2 {
		t.Errorf("Expected default reminder days 2, got %d", config.Alerts.ReminderDays)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.Logging.Level)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("DATABASE_PATH", "/tmp/test.db")
	os.Setenv("ALERTS_REMINDER_DAYS", "3")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("ALERTS_REMINDER_DAYS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify environment variable values
	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Server.Port)
	}

	if config.Server.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got %s", config.Server.Host)
	}

	if config.Database.Path != "/tmp/test.db" {
		t.Errorf("Expected database path '/tmp/test.db', got %s", config.Database.Path)
	}

	if config.Alerts.ReminderDays != 3 {
		t.Errorf("Expected reminder days 3, got %d", config.Alerts.ReminderDays)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", config.Logging.Level)
	}

	if config.Logging.Format != "json" {
		t.Errorf("Expected log format 'json', got %s", config.Logging.Format)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
					Host: "localhost",
				},
				Database: DatabaseConfig{
					Path: "./test.db",
				},
				Alerts: AlertsConfig{
					ReminderDays: 2,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: Config{
				Server: ServerConfig{
					Port: 0,
				},
				Database: DatabaseConfig{
					Path: "./test.db",
				},
				Alerts: AlertsConfig{
					ReminderDays: 2,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: Config{
				Server: ServerConfig{
					Port: 70000,
				},
				Database: DatabaseConfig{
					Path: "./test.db",
				},
				Alerts: AlertsConfig{
					ReminderDays: 2,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "empty database path",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Database: DatabaseConfig{
					Path: "",
				},
				Alerts: AlertsConfig{
					ReminderDays: 2,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "negative reminder days",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Database: DatabaseConfig{
					Path: "./test.db",
				},
				Alerts: AlertsConfig{
					ReminderDays: -1,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Database: DatabaseConfig{
					Path: "./test.db",
				},
				Alerts: AlertsConfig{
					ReminderDays: 2,
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Database: DatabaseConfig{
					Path: "./test.db",
				},
				Alerts: AlertsConfig{
					ReminderDays: 2,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	// Test with valid duration
	os.Setenv("TEST_DURATION", "5m")
	defer os.Unsetenv("TEST_DURATION")

	duration := getEnvAsDuration("TEST_DURATION", time.Minute)
	expected := 5 * time.Minute
	if duration != expected {
		t.Errorf("Expected %v, got %v", expected, duration)
	}

	// Test with invalid duration (should return default)
	os.Setenv("TEST_DURATION_INVALID", "invalid")
	defer os.Unsetenv("TEST_DURATION_INVALID")

	duration = getEnvAsDuration("TEST_DURATION_INVALID", time.Minute)
	expected = time.Minute
	if duration != expected {
		t.Errorf("Expected %v, got %v", expected, duration)
	}

	// Test with missing env var (should return default)
	duration = getEnvAsDuration("MISSING_DURATION", time.Hour)
	expected = time.Hour
	if duration != expected {
		t.Errorf("Expected %v, got %v", expected, duration)
	}
}