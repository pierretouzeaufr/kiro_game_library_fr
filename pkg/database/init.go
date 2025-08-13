package database

import (
	"fmt"
	"log"
)

// Initialize sets up the database with all required tables and indexes
func Initialize(config Config) (*DB, error) {
	// Create database connection
	db, err := NewConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Create migration manager and run migrations
	migrationManager := NewMigrationManager(db)
	if err := migrationManager.Migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Printf("Database initialized successfully at: %s", config.DatabasePath)
	return db, nil
}

// InitializeForTesting creates an in-memory database for testing
func InitializeForTesting() (*DB, error) {
	config := Config{
		DatabasePath: ":memory:",
	}
	
	return Initialize(config)
}