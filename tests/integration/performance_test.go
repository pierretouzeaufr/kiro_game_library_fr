package integration

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"board-game-library/internal/services"
	"board-game-library/pkg/database"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentUserAccess tests the system's ability to handle multiple concurrent users
func TestConcurrentUserAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories and services
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)

	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)

	t.Run("Concurrent User Registration", func(t *testing.T) {
		const numUsers = 100
		var wg sync.WaitGroup
		errors := make(chan error, numUsers)
		users := make(chan *models.User, numUsers)

		start := time.Now()

		// Register users concurrently
		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				user, err := userService.RegisterUser(
					fmt.Sprintf("Concurrent User %d", index),
					fmt.Sprintf("concurrent%d@example.com", index),
				)
				
				if err != nil {
					errors <- err
					return
				}
				users <- user
			}(i)
		}

		wg.Wait()
		close(errors)
		close(users)

		duration := time.Since(start)
		t.Logf("Registered %d users concurrently in %v", numUsers, duration)

		// Check for errors
		var errorCount int
		for err := range errors {
			t.Errorf("Registration error: %v", err)
			errorCount++
		}
		assert.Equal(t, 0, errorCount, "Should have no registration errors")

		// Count successful registrations
		var successCount int
		for range users {
			successCount++
		}
		assert.Equal(t, numUsers, successCount, "All users should be registered successfully")

		// Verify all users exist in database
		allUsers, err := userService.GetAllUsers()
		require.NoError(t, err)
		
		concurrentUsers := 0
		for _, user := range allUsers {
			if len(user.Name) > 15 && user.Name[:15] == "Concurrent User" {
				concurrentUsers++
			}
		}
		assert.Equal(t, numUsers, concurrentUsers)
	})

	t.Run("Concurrent Game Borrowing", func(t *testing.T) {
		const numUsers = 50
		const numGames = 50

		// Create users and games first
		var users []*models.User
		var games []*models.Game

		// Create users
		for i := 0; i < numUsers; i++ {
			user, err := userService.RegisterUser(
				fmt.Sprintf("Borrowing User %d", i),
				fmt.Sprintf("borrowing%d@example.com", i),
			)
			require.NoError(t, err)
			users = append(users, user)
		}

		// Create games
		for i := 0; i < numGames; i++ {
			game, err := gameService.AddGame(
				fmt.Sprintf("Concurrent Game %d", i),
				fmt.Sprintf("Description for game %d", i),
				"Concurrent",
				"good",
			)
			require.NoError(t, err)
			games = append(games, game)
		}

		var wg sync.WaitGroup
		borrowingResults := make(chan error, numUsers)

		start := time.Now()

		// Each user tries to borrow a different game concurrently
		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()
				
				user := users[userIndex]
				game := games[userIndex] // Each user gets a different game
				
				_, err := borrowingService.BorrowGame(
					user.ID,
					game.ID,
					time.Now().Add(14*24*time.Hour),
				)
				
				borrowingResults <- err
			}(i)
		}

		wg.Wait()
		close(borrowingResults)

		duration := time.Since(start)
		t.Logf("Processed %d concurrent borrowings in %v", numUsers, duration)

		// Check results
		var errorCount int
		for err := range borrowingResults {
			if err != nil {
				t.Errorf("Borrowing error: %v", err)
				errorCount++
			}
		}
		assert.Equal(t, 0, errorCount, "Should have no borrowing errors")

		// Verify all games are borrowed
		for _, game := range games {
			updatedGame, err := gameService.GetGame(game.ID)
			require.NoError(t, err)
			assert.False(t, updatedGame.IsAvailable, "Game %d should be borrowed", game.ID)
		}
	})

	t.Run("Concurrent Same Game Borrowing (Race Condition Test)", func(t *testing.T) {
		const numUsers = 10

		// Create one game that multiple users will try to borrow
		game, err := gameService.AddGame("Race Condition Game", "Only one can borrow", "Race", "good")
		require.NoError(t, err)

		// Create multiple users
		var users []*models.User
		for i := 0; i < numUsers; i++ {
			user, err := userService.RegisterUser(
				fmt.Sprintf("Race User %d", i),
				fmt.Sprintf("race%d@example.com", i),
			)
			require.NoError(t, err)
			users = append(users, user)
		}

		var wg sync.WaitGroup
		borrowingResults := make(chan error, numUsers)

		start := time.Now()

		// All users try to borrow the same game concurrently
		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()
				
				user := users[userIndex]
				_, err := borrowingService.BorrowGame(
					user.ID,
					game.ID,
					time.Now().Add(14*24*time.Hour),
				)
				
				borrowingResults <- err
			}(i)
		}

		wg.Wait()
		close(borrowingResults)

		duration := time.Since(start)
		t.Logf("Processed %d concurrent attempts to borrow same game in %v", numUsers, duration)

		// Count successful and failed borrowings
		var successCount, errorCount int
		for err := range borrowingResults {
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}

		// Only one borrowing should succeed
		assert.Equal(t, 1, successCount, "Only one user should successfully borrow the game")
		assert.Equal(t, numUsers-1, errorCount, "All other attempts should fail")

		// Verify game is not available
		updatedGame, err := gameService.GetGame(game.ID)
		require.NoError(t, err)
		assert.False(t, updatedGame.IsAvailable, "Game should be borrowed")
	})
}

// TestLargeDatasetPerformance tests system performance with large amounts of data
func TestLargeDatasetPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories and services
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)
	alertRepo := repositories.NewSQLiteAlertRepository(db)

	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)
	_ = services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	t.Run("Large Dataset Creation", func(t *testing.T) {
		const numUsers = 1000
		const numGames = 2000

		// Create users
		start := time.Now()
		for i := 0; i < numUsers; i++ {
			_, err := userService.RegisterUser(
				fmt.Sprintf("Large Dataset User %d", i),
				fmt.Sprintf("large%d@example.com", i),
			)
			require.NoError(t, err)
		}
		userCreationTime := time.Since(start)
		t.Logf("Created %d users in %v", numUsers, userCreationTime)

		// Create games
		start = time.Now()
		for i := 0; i < numGames; i++ {
			_, err := gameService.AddGame(
				fmt.Sprintf("Large Dataset Game %d", i),
				fmt.Sprintf("Description for large dataset game %d", i),
				"Large",
				"good",
			)
			require.NoError(t, err)
		}
		gameCreationTime := time.Since(start)
		t.Logf("Created %d games in %v", numGames, gameCreationTime)

		// Performance assertions
		assert.Less(t, userCreationTime, 30*time.Second, "User creation should complete within 30 seconds")
		assert.Less(t, gameCreationTime, 60*time.Second, "Game creation should complete within 60 seconds")
	})

	t.Run("Large Dataset Queries", func(t *testing.T) {
		// Test retrieving all users
		start := time.Now()
		users, err := userService.GetAllUsers()
		userQueryTime := time.Since(start)
		require.NoError(t, err)
		t.Logf("Retrieved %d users in %v", len(users), userQueryTime)

		// Test retrieving all games
		start = time.Now()
		games, err := gameService.GetAllGames()
		gameQueryTime := time.Since(start)
		require.NoError(t, err)
		t.Logf("Retrieved %d games in %v", len(games), gameQueryTime)

		// Performance assertions
		assert.Less(t, userQueryTime, 2*time.Second, "User query should complete within 2 seconds")
		assert.Less(t, gameQueryTime, 2*time.Second, "Game query should complete within 2 seconds")
		assert.GreaterOrEqual(t, len(users), 1000, "Should have at least 1000 users")
		assert.GreaterOrEqual(t, len(games), 2000, "Should have at least 2000 games")
	})

	t.Run("Large Dataset Search Performance", func(t *testing.T) {
		// Test game search with large dataset
		start := time.Now()
		searchResults, err := gameService.SearchGames("Large Dataset Game 1")
		searchTime := time.Since(start)
		require.NoError(t, err)
		t.Logf("Search returned %d results in %v", len(searchResults), searchTime)

		// Performance assertions
		assert.Less(t, searchTime, 1*time.Second, "Search should complete within 1 second")
		assert.Greater(t, len(searchResults), 0, "Search should return results")
	})

	t.Run("Large Dataset Borrowing Operations", func(t *testing.T) {
		// Get some users and games for borrowing
		users, err := userService.GetAllUsers()
		require.NoError(t, err)
		require.Greater(t, len(users), 100, "Need at least 100 users for test")

		games, err := gameService.GetAllGames()
		require.NoError(t, err)
		require.Greater(t, len(games), 100, "Need at least 100 games for test")

		// Create 500 borrowings
		const numBorrowings = 500
		start := time.Now()

		for i := 0; i < numBorrowings; i++ {
			user := users[i%len(users)]
			game := games[i] // Use different games to avoid conflicts
			
			_, err := borrowingService.BorrowGame(
				user.ID,
				game.ID,
				time.Now().Add(14*24*time.Hour),
			)
			require.NoError(t, err)
		}

		borrowingTime := time.Since(start)
		t.Logf("Created %d borrowings in %v", numBorrowings, borrowingTime)

		// Performance assertion
		assert.Less(t, borrowingTime, 30*time.Second, "Borrowing operations should complete within 30 seconds")

		// Test querying borrowings
		start = time.Now()
		_, err = userService.GetUserBorrowings(users[0].ID)
		queryTime := time.Since(start)
		require.NoError(t, err)
		t.Logf("Retrieved user borrowings in %v", queryTime)

		assert.Less(t, queryTime, 1*time.Second, "Borrowing query should complete within 1 second")
	})

	t.Run("Memory Usage Monitoring", func(t *testing.T) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Perform memory-intensive operations
		users, err := userService.GetAllUsers()
		require.NoError(t, err)

		games, err := gameService.GetAllGames()
		require.NoError(t, err)

		// Force garbage collection and read memory stats
		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.Alloc - m1.Alloc
		t.Logf("Memory used for loading %d users and %d games: %d bytes", 
			len(users), len(games), memoryUsed)

		// Memory usage should be reasonable (adjust threshold as needed)
		assert.Less(t, memoryUsed, uint64(100*1024*1024), "Memory usage should be less than 100MB")
	})
}

// TestSystemStressTest performs stress testing on the system
func TestSystemStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress tests in short mode")
	}

	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories and services
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)
	alertRepo := repositories.NewSQLiteAlertRepository(db)

	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)
	alertService := services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	t.Run("Mixed Operations Stress Test", func(t *testing.T) {
		const duration = 10 * time.Second
		const numWorkers = 10

		var wg sync.WaitGroup
		stopChan := make(chan struct{})
		operationCounts := make([]int, numWorkers)

		// Start workers performing different operations
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				operationCount := 0
				for {
					select {
					case <-stopChan:
						operationCounts[workerID] = operationCount
						return
					default:
						// Perform random operations
						switch operationCount % 4 {
						case 0:
							// Create user
							_, err := userService.RegisterUser(
								fmt.Sprintf("Stress User %d-%d", workerID, operationCount),
								fmt.Sprintf("stress%d-%d@example.com", workerID, operationCount),
							)
							if err == nil {
								operationCount++
							}
						case 1:
							// Create game
							_, err := gameService.AddGame(
								fmt.Sprintf("Stress Game %d-%d", workerID, operationCount),
								"Stress test game",
								"Stress",
								"good",
							)
							if err == nil {
								operationCount++
							}
						case 2:
							// Query users
							_, err := userService.GetAllUsers()
							if err == nil {
								operationCount++
							}
						case 3:
							// Query games
							_, err := gameService.GetAllGames()
							if err == nil {
								operationCount++
							}
						}
					}
				}
			}(i)
		}

		// Let workers run for specified duration
		time.Sleep(duration)
		close(stopChan)
		wg.Wait()

		// Calculate total operations
		totalOperations := 0
		for i, count := range operationCounts {
			t.Logf("Worker %d performed %d operations", i, count)
			totalOperations += count
		}

		t.Logf("Total operations performed in %v: %d", duration, totalOperations)
		t.Logf("Operations per second: %.2f", float64(totalOperations)/duration.Seconds())

		// Should handle at least some operations per second
		assert.Greater(t, totalOperations, 100, "Should perform at least 100 operations in stress test")
	})

	t.Run("Alert Generation Performance", func(t *testing.T) {
		// Create users and games
		const numUsers = 100
		const numGames = 100

		var users []*models.User
		var games []*models.Game

		for i := 0; i < numUsers; i++ {
			user, err := userService.RegisterUser(
				fmt.Sprintf("Alert User %d", i),
				fmt.Sprintf("alert%d@example.com", i),
			)
			require.NoError(t, err)
			users = append(users, user)
		}

		for i := 0; i < numGames; i++ {
			game, err := gameService.AddGame(
				fmt.Sprintf("Alert Game %d", i),
				"Alert test game",
				"Alert",
				"good",
			)
			require.NoError(t, err)
			games = append(games, game)
		}

		// Create overdue borrowings
		for i := 0; i < numUsers; i++ {
			_, err := borrowingService.BorrowGame(
				users[i].ID,
				games[i].ID,
				time.Now().Add(-24*time.Hour), // Overdue by 1 day
			)
			require.NoError(t, err)
		}

		// Test alert generation performance
		start := time.Now()
		err = alertService.GenerateOverdueAlerts()
		alertGenerationTime := time.Since(start)
		require.NoError(t, err)

		t.Logf("Generated overdue alerts in %v", alertGenerationTime)

		// Verify alerts were created
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		t.Logf("Generated %d alerts", len(alerts))

		// Performance assertion
		assert.Less(t, alertGenerationTime, 5*time.Second, "Alert generation should complete within 5 seconds")
		assert.Equal(t, numUsers, len(alerts), "Should generate one alert per overdue borrowing")
	})
}

// TestMemoryLeakDetection tests for memory leaks during extended operations
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak tests in short mode")
	}

	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories and services
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)

	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)

	t.Run("Memory Usage During Extended Operations", func(t *testing.T) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Perform many operations in a loop
		const iterations = 1000
		for i := 0; i < iterations; i++ {
			// Create user
			user, err := userService.RegisterUser(
				fmt.Sprintf("Memory Test User %d", i),
				fmt.Sprintf("memory%d@test.com", i),
			)
			require.NoError(t, err)

			// Create game
			game, err := gameService.AddGame(
				fmt.Sprintf("Memory Test Game %d", i),
				"Memory test game",
				"Memory",
				"good",
			)
			require.NoError(t, err)

			// Borrow and return
			borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
			require.NoError(t, err)

			err = borrowingService.ReturnGame(borrowing.ID)
			require.NoError(t, err)

			// Periodic garbage collection
			if i%100 == 0 {
				runtime.GC()
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryGrowth := m2.Alloc - m1.Alloc
		t.Logf("Memory growth after %d operations: %d bytes", iterations, memoryGrowth)

		// Memory growth should be reasonable (adjust threshold as needed)
		assert.Less(t, memoryGrowth, uint64(50*1024*1024), "Memory growth should be less than 50MB")
	})

	t.Run("Connection Pool Stress Test", func(t *testing.T) {
		const numConnections = 20
		const operationsPerConnection = 50

		var wg sync.WaitGroup
		errors := make(chan error, numConnections*operationsPerConnection)

		// Start multiple goroutines that perform database operations
		for i := 0; i < numConnections; i++ {
			wg.Add(1)
			go func(connID int) {
				defer wg.Done()

				for j := 0; j < operationsPerConnection; j++ {
					// Perform various database operations
					user := &models.User{
						Name:         fmt.Sprintf("Conn %d User %d", connID, j),
						Email:        fmt.Sprintf("conn%d-user%d@test.com", connID, j),
						RegisteredAt: time.Now(),
						IsActive:     true,
					}

					if err := userRepo.Create(user); err != nil {
						errors <- err
						continue
					}

					// Read operation
					if _, err := userRepo.GetByID(user.ID); err != nil {
						errors <- err
						continue
					}

					// Update operation
					user.Name = fmt.Sprintf("Updated %s", user.Name)
					if err := userRepo.Update(user); err != nil {
						errors <- err
						continue
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Count errors
		var errorCount int
		for err := range errors {
			t.Logf("Connection pool error: %v", err)
			errorCount++
		}

		totalOperations := numConnections * operationsPerConnection * 3 // 3 operations per iteration
		successRate := float64(totalOperations-errorCount) / float64(totalOperations)
		t.Logf("Connection pool success rate: %.2f%% (%d/%d)", successRate*100, totalOperations-errorCount, totalOperations)

		// Should have high success rate
		assert.Greater(t, successRate, 0.95, "Should have >95% success rate under connection stress")
	})
}

// TestScalabilityLimits tests system behavior at scale limits
func TestScalabilityLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability tests in short mode")
	}

	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories and services
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)
	alertRepo := repositories.NewSQLiteAlertRepository(db)

	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)
	alertService := services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	t.Run("Large Scale Data Operations", func(t *testing.T) {
		const numUsers = 2000
		const numGames = 3000
		const numBorrowings = 1000

		// Create large number of users
		start := time.Now()
		users := make([]*models.User, numUsers)
		for i := 0; i < numUsers; i++ {
			user, err := userService.RegisterUser(
				fmt.Sprintf("Scale User %d", i),
				fmt.Sprintf("scale%d@test.com", i),
			)
			require.NoError(t, err)
			users[i] = user
		}
		userCreationTime := time.Since(start)
		t.Logf("Created %d users in %v", numUsers, userCreationTime)

		// Create large number of games
		start = time.Now()
		games := make([]*models.Game, numGames)
		for i := 0; i < numGames; i++ {
			game, err := gameService.AddGame(
				fmt.Sprintf("Scale Game %d", i),
				fmt.Sprintf("Description for scale game %d", i),
				"Scale",
				"good",
			)
			require.NoError(t, err)
			games[i] = game
		}
		gameCreationTime := time.Since(start)
		t.Logf("Created %d games in %v", numGames, gameCreationTime)

		// Create borrowings
		start = time.Now()
		for i := 0; i < numBorrowings; i++ {
			userIndex := i % numUsers
			gameIndex := i % numGames
			
			_, err := borrowingService.BorrowGame(
				users[userIndex].ID,
				games[gameIndex].ID,
				time.Now().Add(14*24*time.Hour),
			)
			require.NoError(t, err)
		}
		borrowingCreationTime := time.Since(start)
		t.Logf("Created %d borrowings in %v", numBorrowings, borrowingCreationTime)

		// Test query performance with large dataset
		start = time.Now()
		allUsers, err := userService.GetAllUsers()
		userQueryTime := time.Since(start)
		require.NoError(t, err)
		assert.Len(t, allUsers, numUsers)
		t.Logf("Queried %d users in %v", len(allUsers), userQueryTime)

		start = time.Now()
		allGames, err := gameService.GetAllGames()
		gameQueryTime := time.Since(start)
		require.NoError(t, err)
		assert.Len(t, allGames, numGames)
		t.Logf("Queried %d games in %v", len(allGames), gameQueryTime)

		// Performance assertions
		assert.Less(t, userCreationTime, 60*time.Second, "User creation should complete within 60 seconds")
		assert.Less(t, gameCreationTime, 90*time.Second, "Game creation should complete within 90 seconds")
		assert.Less(t, borrowingCreationTime, 30*time.Second, "Borrowing creation should complete within 30 seconds")
		assert.Less(t, userQueryTime, 3*time.Second, "User query should complete within 3 seconds")
		assert.Less(t, gameQueryTime, 3*time.Second, "Game query should complete within 3 seconds")
	})

	t.Run("Alert System Scalability", func(t *testing.T) {
		// Create scenario with many overdue items
		const numOverdueItems = 500

		users := make([]*models.User, numOverdueItems)
		games := make([]*models.Game, numOverdueItems)

		// Create users and games
		for i := 0; i < numOverdueItems; i++ {
			user, err := userService.RegisterUser(
				fmt.Sprintf("Alert Scale User %d", i),
				fmt.Sprintf("alertscale%d@test.com", i),
			)
			require.NoError(t, err)
			users[i] = user

			game, err := gameService.AddGame(
				fmt.Sprintf("Alert Scale Game %d", i),
				"Game for alert scalability testing",
				"AlertScale",
				"good",
			)
			require.NoError(t, err)
			games[i] = game
		}

		// Create overdue borrowings
		for i := 0; i < numOverdueItems; i++ {
			borrowing, err := borrowingService.BorrowGame(
				users[i].ID,
				games[i].ID,
				time.Now().Add(-time.Duration(i%10+1)*24*time.Hour), // Various overdue periods
			)
			require.NoError(t, err)

			// Mark as overdue
			borrowing.IsOverdue = true
			err = borrowingRepo.Update(borrowing)
			require.NoError(t, err)
		}

		// Test alert generation performance
		start := time.Now()
		err = alertService.GenerateOverdueAlerts()
		alertGenerationTime := time.Since(start)
		require.NoError(t, err)
		t.Logf("Generated alerts for %d overdue items in %v", numOverdueItems, alertGenerationTime)

		// Verify alerts were created
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, alerts, numOverdueItems)

		// Test alert query performance
		start = time.Now()
		summary, err := alertService.GetAlertsSummaryByUser()
		summaryTime := time.Since(start)
		require.NoError(t, err)
		assert.Len(t, summary, numOverdueItems)
		t.Logf("Generated alert summary for %d users in %v", len(summary), summaryTime)

		// Performance assertions
		assert.Less(t, alertGenerationTime, 10*time.Second, "Alert generation should complete within 10 seconds")
		assert.Less(t, summaryTime, 2*time.Second, "Alert summary should complete within 2 seconds")
	})

	t.Run("Search Performance at Scale", func(t *testing.T) {
		// Test search performance with large dataset
		const numSearchGames = 1000

		// Create games with searchable names
		categories := []string{"Strategy", "Family", "Card", "Board", "Puzzle"}
		for i := 0; i < numSearchGames; i++ {
			category := categories[i%len(categories)]
			_, err := gameService.AddGame(
				fmt.Sprintf("Search %s Game %d", category, i),
				fmt.Sprintf("Searchable game %d in %s category", i, category),
				category,
				"good",
			)
			require.NoError(t, err)
		}

		// Test various search queries
		searchQueries := []string{"Strategy", "Family", "Game", "Search", "Puzzle"}
		
		for _, query := range searchQueries {
			start := time.Now()
			results, err := gameService.SearchGames(query)
			searchTime := time.Since(start)
			require.NoError(t, err)
			
			t.Logf("Search for '%s' returned %d results in %v", query, len(results), searchTime)
			assert.Greater(t, len(results), 0, "Search should return results")
			assert.Less(t, searchTime, 1*time.Second, "Search should complete within 1 second")
		}
	})
}