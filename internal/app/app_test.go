package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if app == nil {
		t.Error("New() returned nil app")
	}

	if app.config == nil {
		t.Error("New() returned app with nil config")
	}

	if app.logger == nil {
		t.Error("New() returned app with nil logger")
	}
}

func TestNewWithInvalidConfig(t *testing.T) {
	// Set invalid port to trigger config validation error
	os.Setenv("SERVER_PORT", "70000")
	defer os.Unsetenv("SERVER_PORT")

	_, err := New()
	if err == nil {
		t.Error("New() expected error for invalid config, got nil")
	}
}

func TestInitialize(t *testing.T) {
	// Use in-memory database for testing
	os.Setenv("DATABASE_PATH", ":memory:")
	defer os.Unsetenv("DATABASE_PATH")

	app, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = app.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if app.db == nil {
		t.Error("Initialize() did not set database connection")
	}

	if app.router == nil {
		t.Error("Initialize() did not set router")
	}

	if app.server == nil {
		t.Error("Initialize() did not set server")
	}

	// Clean up
	app.db.Close()
}

func TestHealthCheckHandler(t *testing.T) {
	// Use in-memory database for testing
	os.Setenv("DATABASE_PATH", ":memory:")
	defer os.Unsetenv("DATABASE_PATH")

	app, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = app.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer app.db.Close()

	// Create test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	app.router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health check returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response contains expected fields
	body := rr.Body.String()
	if !contains(body, "status") {
		t.Error("Health check response missing 'status' field")
	}

	if !contains(body, "message") {
		t.Error("Health check response missing 'message' field")
	}

	if !contains(body, "timestamp") {
		t.Error("Health check response missing 'timestamp' field")
	}
}

func TestStatusHandler(t *testing.T) {
	// Use in-memory database for testing
	os.Setenv("DATABASE_PATH", ":memory:")
	defer os.Unsetenv("DATABASE_PATH")

	app, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = app.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer app.db.Close()

	// Create test request
	req, err := http.NewRequest("GET", "/api/v1/status", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	app.router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status endpoint returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response contains expected fields
	body := rr.Body.String()
	if !contains(body, "database") {
		t.Error("Status response missing 'database' field")
	}

	if !contains(body, "config") {
		t.Error("Status response missing 'config' field")
	}
}

func TestGetters(t *testing.T) {
	// Use in-memory database for testing
	os.Setenv("DATABASE_PATH", ":memory:")
	defer os.Unsetenv("DATABASE_PATH")

	app, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = app.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer app.db.Close()

	// Test GetDB
	db := app.GetDB()
	if db == nil {
		t.Error("GetDB() returned nil")
	}

	// Test GetLogger
	logger := app.GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}

	// Test GetConfig
	config := app.GetConfig()
	if config == nil {
		t.Error("GetConfig() returned nil")
	}
}

func TestShutdown(t *testing.T) {
	// Use in-memory database for testing
	os.Setenv("DATABASE_PATH", ":memory:")
	defer os.Unsetenv("DATABASE_PATH")

	app, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = app.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test shutdown
	err = app.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestInitializeWithInvalidDatabase(t *testing.T) {
	// Set invalid database path
	os.Setenv("DATABASE_PATH", "/invalid/path/that/does/not/exist/test.db")
	defer os.Unsetenv("DATABASE_PATH")

	app, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = app.Initialize()
	if err == nil {
		t.Error("Initialize() expected error for invalid database path, got nil")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}