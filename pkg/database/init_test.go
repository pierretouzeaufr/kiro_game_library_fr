package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := Config{
		DatabasePath: dbPath,
	}

	db, err := Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test that the database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test that all tables exist
	expectedTables := []string{"users", "games", "borrowings", "alerts", "schema_migrations"}
	
	for _, tableName := range expectedTables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query table %s: %v", tableName, err)
		}
		
		if count != 1 {
			t.Errorf("Expected table %s to exist, but it doesn't", tableName)
		}
	}

	// Test that indexes exist
	expectedIndexes := []string{
		"idx_borrowings_user_id",
		"idx_borrowings_game_id", 
		"idx_borrowings_due_date",
		"idx_alerts_user_id",
		"idx_alerts_is_read",
	}
	
	for _, indexName := range expectedIndexes {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query index %s: %v", indexName, err)
		}
		
		if count != 1 {
			t.Errorf("Expected index %s to exist, but it doesn't", indexName)
		}
	}
}

func TestInitializeForTesting(t *testing.T) {
	db, err := InitializeForTesting()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Test that we can ping the database
	if err := db.Ping(); err != nil {
		t.Errorf("Failed to ping test database: %v", err)
	}

	// Test that all tables exist
	expectedTables := []string{"users", "games", "borrowings", "alerts", "schema_migrations"}
	
	for _, tableName := range expectedTables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query table %s: %v", tableName, err)
		}
		
		if count != 1 {
			t.Errorf("Expected table %s to exist, but it doesn't", tableName)
		}
	}
}