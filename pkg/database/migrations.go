package database

import (
	"fmt"
	"sort"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db         *DB
	migrations []Migration
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DB) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: getInitialMigrations(),
	}
}

// Initialize sets up the migration tracking table
func (mm *MigrationManager) Initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := mm.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// Migrate runs all pending migrations
func (mm *MigrationManager) Migrate() error {
	if err := mm.Initialize(); err != nil {
		return err
	}

	appliedVersions, err := mm.getAppliedVersions()
	if err != nil {
		return err
	}

	// Sort migrations by version
	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version < mm.migrations[j].Version
	})

	for _, migration := range mm.migrations {
		if _, applied := appliedVersions[migration.Version]; applied {
			continue
		}

		if err := mm.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", 
				migration.Version, migration.Name, err)
		}
	}

	return nil
}

// getAppliedVersions returns a map of applied migration versions
func (mm *MigrationManager) getAppliedVersions() (map[int]bool, error) {
	query := "SELECT version FROM schema_migrations"
	rows, err := mm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = true
	}

	return versions, rows.Err()
}

// applyMigration applies a single migration
func (mm *MigrationManager) applyMigration(migration Migration) error {
	tx, err := mm.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration statements
	statements := strings.Split(migration.Up, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement '%s': %w", stmt, err)
		}
	}

	// Record migration as applied
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
		migration.Version, migration.Name,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}