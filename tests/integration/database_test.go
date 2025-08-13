package integration

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"board-game-library/pkg/database"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseIntegration tests database operations with real SQLite database
func TestDatabaseIntegration(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "integration_test.db")

	// Initialize database with file storage
	config := database.Config{
		DatabasePath: dbPath,
	}

	db, err := database.Initialize(config)
	require.NoError(t, err)
	defer db.Close()

	// Verify database file was created
	_, err = os.Stat(dbPath)
	require.NoError(t, err, "Database file should exist")

	t.Run("Database Schema Integrity", func(t *testing.T) {
		// Test that all required tables exist
		expectedTables := []string{"users", "games", "borrowings", "alerts", "schema_migrations"}
		
		for _, tableName := range expectedTables {
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count, "Table %s should exist", tableName)
		}

		// Test that all required indexes exist
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
			require.NoError(t, err)
			assert.Equal(t, 1, count, "Index %s should exist", indexName)
		}
	})

	t.Run("Foreign Key Constraints", func(t *testing.T) {
		userRepo := repositories.NewSQLiteUserRepository(db)
		gameRepo := repositories.NewSQLiteGameRepository(db)
		borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)

		// Create user and game
		user := &models.User{
			Name:         "FK Test User",
			Email:        "fk@example.com",
			RegisteredAt: time.Now(),
			IsActive:     true,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		game := &models.Game{
			Name:        "FK Test Game",
			Description: "A game for FK testing",
			Category:    "Test",
			EntryDate:   time.Now(),
			Condition:   "good",
			IsAvailable: true,
		}
		err = gameRepo.Create(game)
		require.NoError(t, err)

		// Test valid foreign key references
		borrowing := &models.Borrowing{
			UserID:     user.ID,
			GameID:     game.ID,
			BorrowedAt: time.Now(),
			DueDate:    time.Now().Add(14 * 24 * time.Hour),
			IsOverdue:  false,
		}
		err = borrowingRepo.Create(borrowing)
		require.NoError(t, err)

		// Test invalid foreign key reference (should fail)
		invalidBorrowing := &models.Borrowing{
			UserID:     99999, // Non-existent user
			GameID:     game.ID,
			BorrowedAt: time.Now(),
			DueDate:    time.Now().Add(14 * 24 * time.Hour),
			IsOverdue:  false,
		}
		err = borrowingRepo.Create(invalidBorrowing)
		assert.Error(t, err, "Should fail due to foreign key constraint")
	})

	t.Run("Transaction Integrity", func(t *testing.T) {
		userRepo := repositories.NewSQLiteUserRepository(db)

		// Start a transaction
		tx, err := db.Begin()
		require.NoError(t, err)

		// Create user within transaction
		user := &models.User{
			Name:         "Transaction Test User",
			Email:        "transaction@example.com",
			RegisteredAt: time.Now(),
			IsActive:     true,
		}

		// Insert user in transaction
		_, err = tx.Exec(`
			INSERT INTO users (name, email, registered_at, is_active) 
			VALUES (?, ?, ?, ?)`,
			user.Name, user.Email, user.RegisteredAt, user.IsActive)
		require.NoError(t, err)

		// Rollback transaction
		err = tx.Rollback()
		require.NoError(t, err)

		// Verify user was not created (due to rollback)
		_, err = userRepo.GetByEmail("transaction@example.com")
		assert.Error(t, err, "User should not exist after rollback")

		// Now test successful transaction
		tx, err = db.Begin()
		require.NoError(t, err)

		result, err := tx.Exec(`
			INSERT INTO users (name, email, registered_at, is_active) 
			VALUES (?, ?, ?, ?)`,
			user.Name, user.Email, user.RegisteredAt, user.IsActive)
		require.NoError(t, err)

		// Commit transaction
		err = tx.Commit()
		require.NoError(t, err)

		// Verify user was created
		id, err := result.LastInsertId()
		require.NoError(t, err)
		
		createdUser, err := userRepo.GetByID(int(id))
		require.NoError(t, err)
		assert.Equal(t, user.Name, createdUser.Name)
		assert.Equal(t, user.Email, createdUser.Email)
	})

	t.Run("Database Connection Pool", func(t *testing.T) {
		// Test multiple concurrent connections
		userRepo := repositories.NewSQLiteUserRepository(db)

		// Create multiple users concurrently
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func(index int) {
				defer func() { done <- true }()
				
				user := &models.User{
					Name:         fmt.Sprintf("Concurrent User %d", index),
					Email:        fmt.Sprintf("concurrent%d@example.com", index),
					RegisteredAt: time.Now(),
					IsActive:     true,
				}
				
				err := userRepo.Create(user)
				assert.NoError(t, err)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		// Verify all users were created
		users, err := userRepo.GetAll()
		require.NoError(t, err)
		
		concurrentUsers := 0
		for _, user := range users {
			if strings.Contains(user.Name, "Concurrent User") {
				concurrentUsers++
			}
		}
		assert.Equal(t, 5, concurrentUsers)
	})

	t.Run("Database Persistence", func(t *testing.T) {
		userRepo := repositories.NewSQLiteUserRepository(db)

		// Create a user
		user := &models.User{
			Name:         "Persistence Test User",
			Email:        "persistence@example.com",
			RegisteredAt: time.Now(),
			IsActive:     true,
		}
		err := userRepo.Create(user)
		require.NoError(t, err)

		// Close the database connection
		err = db.Close()
		require.NoError(t, err)

		// Reopen the database
		db2, err := database.Initialize(config)
		require.NoError(t, err)
		defer db2.Close()

		// Verify the user still exists
		userRepo2 := repositories.NewSQLiteUserRepository(db2)
		retrievedUser, err := userRepo2.GetByEmail("persistence@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.Name, retrievedUser.Name)
		assert.Equal(t, user.Email, retrievedUser.Email)
	})
}

// TestDatabaseMigrations tests the database migration system
func TestDatabaseMigrations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migration_test.db")

	config := database.Config{
		DatabasePath: dbPath,
	}

	t.Run("Initial Migration", func(t *testing.T) {
		db, err := database.Initialize(config)
		require.NoError(t, err)
		defer db.Close()

		// Check that schema_migrations table exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Check that initial migration was recorded
		var migrationCount int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
		require.NoError(t, err)
		assert.Greater(t, migrationCount, 0, "Should have at least one migration recorded")
	})

	t.Run("Migration Idempotency", func(t *testing.T) {
		// Initialize database twice - should not cause errors
		db1, err := database.Initialize(config)
		require.NoError(t, err)
		db1.Close()

		db2, err := database.Initialize(config)
		require.NoError(t, err)
		defer db2.Close()

		// Verify database is still functional
		var count int
		err = db2.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		require.NoError(t, err)
	})
}

// TestDatabasePerformance tests database performance characteristics
func TestDatabasePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)

	t.Run("Bulk Insert Performance", func(t *testing.T) {
		start := time.Now()

		// Insert 1000 users
		for i := 0; i < 1000; i++ {
			user := &models.User{
				Name:         fmt.Sprintf("Bulk User %d", i),
				Email:        fmt.Sprintf("bulk%d@example.com", i),
				RegisteredAt: time.Now(),
				IsActive:     true,
			}
			err := userRepo.Create(user)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		t.Logf("Inserted 1000 users in %v", duration)
		
		// Should complete within reasonable time (adjust threshold as needed)
		assert.Less(t, duration, 10*time.Second, "Bulk insert should complete within 10 seconds")
	})

	t.Run("Query Performance with Large Dataset", func(t *testing.T) {
		// Insert 500 games
		for i := 0; i < 500; i++ {
			game := &models.Game{
				Name:        fmt.Sprintf("Performance Game %d", i),
				Description: fmt.Sprintf("Description for game %d", i),
				Category:    "Performance",
				EntryDate:   time.Now(),
				Condition:   "good",
				IsAvailable: true,
			}
			err := gameRepo.Create(game)
			require.NoError(t, err)
		}

		// Test query performance
		start := time.Now()
		games, err := gameRepo.GetAll()
		duration := time.Since(start)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(games), 500)
		t.Logf("Retrieved %d games in %v", len(games), duration)
		
		// Should complete within reasonable time
		assert.Less(t, duration, 1*time.Second, "Query should complete within 1 second")
	})

	t.Run("Search Performance", func(t *testing.T) {
		start := time.Now()
		games, err := gameRepo.Search("Performance Game")
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Greater(t, len(games), 0)
		t.Logf("Search returned %d games in %v", len(games), duration)
		
		// Search should be fast
		assert.Less(t, duration, 500*time.Millisecond, "Search should complete within 500ms")
	})
}

// TestDatabaseConcurrency tests database operations under concurrent access
func TestDatabaseConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency tests in short mode")
	}

	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	_ = repositories.NewSQLiteBorrowingRepository(db)

	t.Run("Concurrent User Creation", func(t *testing.T) {
		const numWorkers = 10
		const usersPerWorker = 10

		done := make(chan error, numWorkers)

		// Start concurrent workers
		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				var err error
				defer func() { done <- err }()

				for j := 0; j < usersPerWorker; j++ {
					user := &models.User{
						Name:         fmt.Sprintf("Concurrent User %d-%d", workerID, j),
						Email:        fmt.Sprintf("concurrent%d-%d@test.com", workerID, j),
						RegisteredAt: time.Now(),
						IsActive:     true,
					}
					
					if createErr := userRepo.Create(user); createErr != nil {
						err = createErr
						return
					}
				}
			}(i)
		}

		// Wait for all workers to complete
		var errorCount int
		for i := 0; i < numWorkers; i++ {
			if err := <-done; err != nil {
				t.Errorf("Worker error: %v", err)
				errorCount++
			}
		}

		assert.Equal(t, 0, errorCount, "Should have no errors in concurrent user creation")

		// Verify all users were created
		users, err := userRepo.GetAll()
		require.NoError(t, err)
		
		concurrentUsers := 0
		for _, user := range users {
			if strings.Contains(user.Name, "Concurrent User") {
				concurrentUsers++
			}
		}
		assert.Equal(t, numWorkers*usersPerWorker, concurrentUsers)
	})

	t.Run("Concurrent Game Borrowing Race Condition", func(t *testing.T) {
		// Create one game that multiple users will try to borrow simultaneously
		game := &models.Game{
			Name:        "Race Condition Game",
			Description: "Game for testing race conditions",
			Category:    "Test",
			EntryDate:   time.Now(),
			Condition:   "good",
			IsAvailable: true,
		}
		err := gameRepo.Create(game)
		require.NoError(t, err)

		// Create multiple users
		const numUsers = 5
		users := make([]*models.User, numUsers)
		for i := 0; i < numUsers; i++ {
			user := &models.User{
				Name:         fmt.Sprintf("Race User %d", i),
				Email:        fmt.Sprintf("race%d@test.com", i),
				RegisteredAt: time.Now(),
				IsActive:     true,
			}
			err := userRepo.Create(user)
			require.NoError(t, err)
			users[i] = user
		}

		// All users try to borrow the same game simultaneously
		results := make(chan error, numUsers)
		
		for i := 0; i < numUsers; i++ {
			go func(userIndex int) {
				borrowing := &models.Borrowing{
					UserID:     users[userIndex].ID,
					GameID:     game.ID,
					BorrowedAt: time.Now(),
					DueDate:    time.Now().Add(14 * 24 * time.Hour),
					IsOverdue:  false,
				}
				
				// Try to create borrowing and update game availability
				tx, err := db.Begin()
				if err != nil {
					results <- err
					return
				}
				defer tx.Rollback()

				// Check if game is still available
				var isAvailable bool
				err = tx.QueryRow("SELECT is_available FROM games WHERE id = ?", game.ID).Scan(&isAvailable)
				if err != nil {
					results <- err
					return
				}

				if !isAvailable {
					results <- fmt.Errorf("game not available")
					return
				}

				// Create borrowing
				_, err = tx.Exec(`
					INSERT INTO borrowings (user_id, game_id, borrowed_at, due_date, is_overdue) 
					VALUES (?, ?, ?, ?, ?)`,
					borrowing.UserID, borrowing.GameID, borrowing.BorrowedAt, borrowing.DueDate, borrowing.IsOverdue)
				if err != nil {
					results <- err
					return
				}

				// Update game availability
				_, err = tx.Exec("UPDATE games SET is_available = ? WHERE id = ?", false, game.ID)
				if err != nil {
					results <- err
					return
				}

				results <- tx.Commit()
			}(i)
		}

		// Collect results
		var successCount, errorCount int
		for i := 0; i < numUsers; i++ {
			if err := <-results; err != nil {
				errorCount++
			} else {
				successCount++
			}
		}

		// Only one should succeed due to race condition handling
		assert.Equal(t, 1, successCount, "Only one user should successfully borrow the game")
		assert.Equal(t, numUsers-1, errorCount, "All other attempts should fail")

		// Verify game is not available
		updatedGame, err := gameRepo.GetByID(game.ID)
		require.NoError(t, err)
		assert.False(t, updatedGame.IsAvailable)
	})

	t.Run("Database Lock Handling", func(t *testing.T) {
		// Test database lock handling under high contention
		const numOperations = 50
		results := make(chan error, numOperations)

		// Perform mixed read/write operations concurrently
		for i := 0; i < numOperations; i++ {
			go func(opIndex int) {
				var err error
				defer func() { results <- err }()

				switch opIndex % 4 {
				case 0:
					// Create user
					user := &models.User{
						Name:         fmt.Sprintf("Lock Test User %d", opIndex),
						Email:        fmt.Sprintf("lock%d@test.com", opIndex),
						RegisteredAt: time.Now(),
						IsActive:     true,
					}
					err = userRepo.Create(user)
				case 1:
					// Read all users
					_, err = userRepo.GetAll()
				case 2:
					// Create game
					game := &models.Game{
						Name:        fmt.Sprintf("Lock Test Game %d", opIndex),
						Description: "Game for lock testing",
						Category:    "Test",
						EntryDate:   time.Now(),
						Condition:   "good",
						IsAvailable: true,
					}
					err = gameRepo.Create(game)
				case 3:
					// Read all games
					_, err = gameRepo.GetAll()
				}
			}(i)
		}

		// Wait for all operations to complete
		var errorCount int
		for i := 0; i < numOperations; i++ {
			if err := <-results; err != nil {
				t.Logf("Operation error: %v", err)
				errorCount++
			}
		}

		// Should handle concurrent operations gracefully
		assert.Less(t, errorCount, numOperations/2, "Should handle most concurrent operations successfully")
	})
}

// TestDatabaseRecovery tests database recovery scenarios
func TestDatabaseRecovery(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "recovery_test.db")

	config := database.Config{
		DatabasePath: dbPath,
	}

	t.Run("Database Recovery After Corruption Simulation", func(t *testing.T) {
		// Create initial database
		db1, err := database.Initialize(config)
		require.NoError(t, err)

		userRepo := repositories.NewSQLiteUserRepository(db1)

		// Add some data
		user := &models.User{
			Name:         "Recovery Test User",
			Email:        "recovery@test.com",
			RegisteredAt: time.Now(),
			IsActive:     true,
		}
		err = userRepo.Create(user)
		require.NoError(t, err)

		// Close database properly
		err = db1.Close()
		require.NoError(t, err)

		// Reopen database and verify data integrity
		db2, err := database.Initialize(config)
		require.NoError(t, err)
		defer db2.Close()

		userRepo2 := repositories.NewSQLiteUserRepository(db2)
		users, err := userRepo2.GetAll()
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "Recovery Test User", users[0].Name)
	})

	t.Run("Database Schema Validation", func(t *testing.T) {
		db, err := database.Initialize(config)
		require.NoError(t, err)
		defer db.Close()

		// Verify all required tables exist with correct structure
		tables := []string{"users", "games", "borrowings", "alerts"}
		
		for _, tableName := range tables {
			// Check table exists
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count, "Table %s should exist", tableName)

			// Check table has data (can be queried)
			_, err = db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName))
			require.NoError(t, err, "Should be able to query table %s", tableName)
		}

		// Verify foreign key constraints are enabled
		var fkEnabled int
		err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
		require.NoError(t, err)
		assert.Equal(t, 1, fkEnabled, "Foreign keys should be enabled")
	})
}

// TestDatabaseBackupAndRestore tests backup and restore functionality
func TestDatabaseBackupAndRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping backup/restore tests in short mode")
	}

	tempDir := t.TempDir()
	originalDBPath := filepath.Join(tempDir, "original.db")
	backupDBPath := filepath.Join(tempDir, "backup.db")

	originalConfig := database.Config{DatabasePath: originalDBPath}
	backupConfig := database.Config{DatabasePath: backupDBPath}

	t.Run("Database Backup and Restore", func(t *testing.T) {
		// Create original database with data
		originalDB, err := database.Initialize(originalConfig)
		require.NoError(t, err)

		userRepo := repositories.NewSQLiteUserRepository(originalDB)
		gameRepo := repositories.NewSQLiteGameRepository(originalDB)

		// Add test data
		user := &models.User{
			Name:         "Backup Test User",
			Email:        "backup@test.com",
			RegisteredAt: time.Now(),
			IsActive:     true,
		}
		err = userRepo.Create(user)
		require.NoError(t, err)

		game := &models.Game{
			Name:        "Backup Test Game",
			Description: "Game for backup testing",
			Category:    "Test",
			EntryDate:   time.Now(),
			Condition:   "good",
			IsAvailable: true,
		}
		err = gameRepo.Create(game)
		require.NoError(t, err)

		originalDB.Close()

		// Copy database file (simulate backup)
		originalData, err := os.ReadFile(originalDBPath)
		require.NoError(t, err)
		
		err = os.WriteFile(backupDBPath, originalData, 0644)
		require.NoError(t, err)

		// Open backup database and verify data
		backupDB, err := database.Initialize(backupConfig)
		require.NoError(t, err)
		defer backupDB.Close()

		backupUserRepo := repositories.NewSQLiteUserRepository(backupDB)
		backupGameRepo := repositories.NewSQLiteGameRepository(backupDB)

		// Verify users
		users, err := backupUserRepo.GetAll()
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "Backup Test User", users[0].Name)

		// Verify games
		games, err := backupGameRepo.GetAll()
		require.NoError(t, err)
		assert.Len(t, games, 1)
		assert.Equal(t, "Backup Test Game", games[0].Name)
	})
}