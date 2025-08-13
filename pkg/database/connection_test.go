package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConnection(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := Config{
		DatabasePath: dbPath,
	}

	db, err := NewConnection(config)
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	// Test that the database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test that we can ping the database
	if err := db.Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestNewConnectionInMemory(t *testing.T) {
	config := Config{
		DatabasePath: ":memory:",
	}

	db, err := NewConnection(config)
	if err != nil {
		t.Fatalf("Failed to create in-memory database connection: %v", err)
	}
	defer db.Close()

	// Test that we can ping the database
	if err := db.Ping(); err != nil {
		t.Errorf("Failed to ping in-memory database: %v", err)
	}
}

func TestNewConnectionInvalidPath(t *testing.T) {
	config := Config{
		DatabasePath: "/invalid/path/that/does/not/exist/test.db",
	}

	_, err := NewConnection(config)
	if err == nil {
		t.Error("Expected error for invalid database path, but got none")
	}
}