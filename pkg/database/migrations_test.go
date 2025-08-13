package database

import (
	"testing"
)

func TestMigrationManager_Initialize(t *testing.T) {
	db, err := InitializeForTesting()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	mm := NewMigrationManager(db)
	
	if err := mm.Initialize(); err != nil {
		t.Fatalf("Failed to initialize migration manager: %v", err)
	}

	// Check that schema_migrations table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query schema_migrations table: %v", err)
	}
	
	if count != 1 {
		t.Errorf("Expected schema_migrations table to exist, but it doesn't")
	}
}

func TestMigrationManager_Migrate(t *testing.T) {
	db, err := InitializeForTesting()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	mm := NewMigrationManager(db)
	
	if err := mm.Migrate(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Check that all expected tables exist
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

	// Check that all migrations were recorded
	var migrationCount int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
	if err != nil {
		t.Fatalf("Failed to count migrations: %v", err)
	}
	
	expectedMigrations := len(getInitialMigrations())
	if migrationCount != expectedMigrations {
		t.Errorf("Expected %d migrations to be recorded, but got %d", expectedMigrations, migrationCount)
	}
}

func TestMigrationManager_MigrateIdempotent(t *testing.T) {
	db, err := InitializeForTesting()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	mm := NewMigrationManager(db)
	
	// Run migrations twice
	if err := mm.Migrate(); err != nil {
		t.Fatalf("Failed to run migrations first time: %v", err)
	}
	
	if err := mm.Migrate(); err != nil {
		t.Fatalf("Failed to run migrations second time: %v", err)
	}

	// Check that migrations were not duplicated
	var migrationCount int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
	if err != nil {
		t.Fatalf("Failed to count migrations: %v", err)
	}
	
	expectedMigrations := len(getInitialMigrations())
	if migrationCount != expectedMigrations {
		t.Errorf("Expected %d migrations to be recorded after running twice, but got %d", expectedMigrations, migrationCount)
	}
}