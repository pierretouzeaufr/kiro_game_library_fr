package integration

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"board-game-library/internal/services"
	"board-game-library/pkg/database"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteLibraryLifecycle tests a complete library management lifecycle
func TestCompleteLibraryLifecycle(t *testing.T) {
	// Setup test database
	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)
	alertRepo := repositories.NewSQLiteAlertRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)
	alertService := services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	t.Run("Complete Library Management Lifecycle", func(t *testing.T) {
		// Phase 1: Library Initialization
		t.Log("Phase 1: Setting up library with users and games")
		
		// Register library members
		_, err := userService.RegisterUser("Alice Librarian", "alice@library.com")
		require.NoError(t, err)
		
		members := make([]*models.User, 5)
		memberNames := []string{"Bob Reader", "Carol Gamer", "Dave Player", "Eve Collector", "Frank Enthusiast"}
		memberEmails := []string{"bob@email.com", "carol@email.com", "dave@email.com", "eve@email.com", "frank@email.com"}
		for i, name := range memberNames {
			member, err := userService.RegisterUser(name, memberEmails[i])
			require.NoError(t, err)
			members[i] = member
		}

		// Add diverse game collection
		games := make([]*models.Game, 10)
		gameData := []struct {
			name, description, category, condition string
		}{
			{"Monopoly", "Classic property trading game", "Family", "good"},
			{"Settlers of Catan", "Resource management and trading", "Strategy", "excellent"},
			{"Scrabble", "Word formation board game", "Word", "good"},
			{"Risk", "Global domination strategy game", "Strategy", "fair"},
			{"Clue", "Mystery solving deduction game", "Mystery", "good"},
			{"Chess", "Classic strategy game for two players", "Strategy", "excellent"},
			{"Checkers", "Traditional board game", "Strategy", "good"},
			{"Trivial Pursuit", "General knowledge quiz game", "Trivia", "fair"},
			{"Pictionary", "Drawing and guessing game", "Party", "good"},
			{"Yahtzee", "Dice rolling game", "Dice", "excellent"},
		}

		for i, data := range gameData {
			game, err := gameService.AddGame(data.name, data.description, data.category, data.condition)
			require.NoError(t, err)
			games[i] = game
		}

		// Verify initial state
		allUsers, err := userService.GetAllUsers()
		require.NoError(t, err)
		assert.Len(t, allUsers, 6) // 1 librarian + 5 members

		allGames, err := gameService.GetAllGames()
		require.NoError(t, err)
		assert.Len(t, allGames, 10)

		availableGames, err := gameService.GetAvailableGames()
		require.NoError(t, err)
		assert.Len(t, availableGames, 10) // All games should be available initially

		// Phase 2: Active Borrowing Period
		t.Log("Phase 2: Active borrowing period with multiple users")

		// Bob borrows strategy games
		bobBorrowing1, err := borrowingService.BorrowGame(members[0].ID, games[1].ID, time.Now().Add(14*24*time.Hour)) // Catan
		require.NoError(t, err)
		
		_, err = borrowingService.BorrowGame(members[0].ID, games[3].ID, time.Now().Add(14*24*time.Hour)) // Risk
		require.NoError(t, err)

		// Carol borrows family games
		carolBorrowing, err := borrowingService.BorrowGame(members[1].ID, games[0].ID, time.Now().Add(14*24*time.Hour)) // Monopoly
		require.NoError(t, err)

		// Dave borrows word games
		_, err = borrowingService.BorrowGame(members[2].ID, games[2].ID, time.Now().Add(14*24*time.Hour)) // Scrabble
		require.NoError(t, err)

		// Verify borrowing state
		bobActiveBorrowings, err := userService.GetActiveUserBorrowings(members[0].ID)
		require.NoError(t, err)
		assert.Len(t, bobActiveBorrowings, 2)

		// Verify game availability
		availableGames, err = gameService.GetAvailableGames()
		require.NoError(t, err)
		assert.Len(t, availableGames, 6) // 4 games borrowed, 6 available

		// Test search functionality during active period
		strategyGames, err := gameService.SearchGames("Strategy")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(strategyGames), 2) // Should find strategy games (some borrowed, some available)

		// Phase 3: Due Date Management and Extensions
		t.Log("Phase 3: Due date management and extensions")

		// Extend Bob's Catan borrowing
		newDueDate := time.Now().Add(21 * 24 * time.Hour)
		err = borrowingService.ExtendDueDate(bobBorrowing1.ID, newDueDate)
		require.NoError(t, err)

		// Verify extension
		extendedBorrowing, err := borrowingService.GetBorrowingDetails(bobBorrowing1.ID)
		require.NoError(t, err)
		assert.True(t, extendedBorrowing.DueDate.After(time.Now().Add(20*24*time.Hour)))

		// Phase 4: Returns and New Borrowings
		t.Log("Phase 4: Processing returns and new borrowings")

		// Carol returns Monopoly
		err = borrowingService.ReturnGame(carolBorrowing.ID)
		require.NoError(t, err)

		// Verify Monopoly is available again
		monopolyGame, err := gameService.GetGame(games[0].ID)
		require.NoError(t, err)
		assert.True(t, monopolyGame.IsAvailable)

		// Eve borrows the returned Monopoly
		eveBorrowing, err := borrowingService.BorrowGame(members[3].ID, games[0].ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		assert.NotZero(t, eveBorrowing.ID)

		// Frank tries to borrow an unavailable game (should fail)
		_, err = borrowingService.BorrowGame(members[4].ID, games[1].ID, time.Now().Add(14*24*time.Hour)) // Catan (borrowed by Bob)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")

		// Phase 5: Overdue Scenario
		t.Log("Phase 5: Handling overdue items and alerts")

		// Create an overdue scenario by creating a normal borrowing then backdating it
		overdueBorrowing, err := borrowingService.BorrowGame(members[4].ID, games[4].ID, time.Now().Add(14*24*time.Hour)) // Clue, initially normal
		require.NoError(t, err)
		
		// Manually update to be overdue (simulating time passage)
		overdueBorrowing.DueDate = time.Now().Add(-3 * 24 * time.Hour) // 3 days overdue

		// Mark as overdue (simulating time passage)
		overdueBorrowing.IsOverdue = true
		err = borrowingRepo.Update(overdueBorrowing)
		require.NoError(t, err)

		// Generate overdue alerts
		err = alertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		// Verify alert was created
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, "overdue", alerts[0].Type)
		assert.Equal(t, members[4].ID, alerts[0].UserID) // Frank has overdue item

		// Verify Frank cannot borrow while having overdue items
		canBorrow, err := userService.CanUserBorrow(members[4].ID)
		require.Error(t, err)
		assert.False(t, canBorrow)

		// Phase 6: Reminder System
		t.Log("Phase 6: Testing reminder system")

		// Create a borrowing due soon (1 day)
		_, err = borrowingService.BorrowGame(members[3].ID, games[5].ID, time.Now().Add(1*24*time.Hour)) // Eve borrows Chess
		require.NoError(t, err)

		// Generate reminder alerts
		err = alertService.GenerateReminderAlerts()
		require.NoError(t, err)

		// Verify reminder alert was created
		allAlerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		
		var reminderAlert *models.Alert
		for _, alert := range allAlerts {
			if alert.Type == "reminder" && alert.UserID == members[3].ID {
				reminderAlert = alert
				break
			}
		}
		require.NotNil(t, reminderAlert)
		assert.Contains(t, reminderAlert.Message, "due")

		// Phase 7: Alert Management
		t.Log("Phase 7: Managing alerts and notifications")

		// Get alert summary
		summary, err := alertService.GetAlertsSummaryByUser()
		require.NoError(t, err)
		assert.Len(t, summary, 2) // Frank (overdue) and Eve (reminder)

		// Mark Frank's overdue alert as read
		frankSummary := summary[members[4].ID]
		err = alertService.MarkAlertAsRead(frankSummary.Alerts[0].ID)
		require.NoError(t, err)

		// Verify alert is marked as read
		updatedAlerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		for _, alert := range updatedAlerts {
			if alert.ID == frankSummary.Alerts[0].ID {
				assert.True(t, alert.IsRead)
			}
		}

		// Phase 8: Resolution and Cleanup
		t.Log("Phase 8: Resolving overdue items and final cleanup")

		// Frank returns overdue game
		err = borrowingService.ReturnGame(overdueBorrowing.ID)
		require.NoError(t, err)

		// Verify Frank can now borrow again
		canBorrow, err = userService.CanUserBorrow(members[4].ID)
		require.NoError(t, err)
		assert.True(t, canBorrow)

		// Frank borrows an available game
		frankBorrowing, err := borrowingService.BorrowGame(members[4].ID, games[6].ID, time.Now().Add(14*24*time.Hour)) // Checkers
		require.NoError(t, err)
		assert.NotZero(t, frankBorrowing.ID)

		// Phase 9: Final State Verification
		t.Log("Phase 9: Final state verification and reporting")

		// Check final game availability
		finalAvailableGames, err := gameService.GetAvailableGames()
		require.NoError(t, err)
		t.Logf("Final available games: %d", len(finalAvailableGames))

		// Check all user borrowing histories
		for i, member := range members {
			history, err := userService.GetUserBorrowings(member.ID)
			require.NoError(t, err)
			t.Logf("Member %d (%s) borrowing history: %d items", i+1, member.Name, len(history))
		}

		// Verify all active borrowings
		totalActiveBorrowings := 0
		for _, member := range members {
			activeBorrowings, err := userService.GetActiveUserBorrowings(member.ID)
			require.NoError(t, err)
			totalActiveBorrowings += len(activeBorrowings)
		}
		t.Logf("Total active borrowings: %d", totalActiveBorrowings)

		// Verify data consistency
		allBorrowings, err := borrowingRepo.GetAll()
		require.NoError(t, err)
		t.Logf("Total borrowings in system: %d", len(allBorrowings))

		// Count returned vs active borrowings
		var returnedCount, activeCount int
		for _, borrowing := range allBorrowings {
			if borrowing.ReturnedAt != nil {
				returnedCount++
			} else {
				activeCount++
			}
		}
		t.Logf("Returned borrowings: %d, Active borrowings: %d", returnedCount, activeCount)
		assert.Equal(t, totalActiveBorrowings, activeCount, "Active borrowing counts should match")

		// Final assertions
		assert.Greater(t, len(allBorrowings), 5, "Should have multiple borrowings recorded")
		assert.Greater(t, returnedCount, 0, "Should have some returned items")
		assert.Greater(t, activeCount, 0, "Should have some active borrowings")
		
		// Verify all users can be retrieved
		finalUsers, err := userService.GetAllUsers()
		require.NoError(t, err)
		assert.Len(t, finalUsers, 6)

		// Verify all games can be retrieved
		finalGames, err := gameService.GetAllGames()
		require.NoError(t, err)
		assert.Len(t, finalGames, 10)
	})
}

// TestBusinessRuleEnforcement tests that business rules are properly enforced
func TestBusinessRuleEnforcement(t *testing.T) {
	// Setup test database
	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)

	t.Run("Business Rule Enforcement", func(t *testing.T) {
		// Create test data
		user, err := userService.RegisterUser("Rule Test User", "rules@test.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Rule Test Game", "Game for testing rules", "Test", "good")
		require.NoError(t, err)

		// Rule 1: Cannot borrow unavailable game
		borrowing1, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// Try to borrow same game again (should fail)
		_, err = borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")

		// Rule 2: Cannot return already returned item
		err = borrowingService.ReturnGame(borrowing1.ID)
		require.NoError(t, err)

		// Try to return again (should fail)
		err = borrowingService.ReturnGame(borrowing1.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already been returned")

		// Rule 3: Cannot extend due date for returned item
		err = borrowingService.ExtendDueDate(borrowing1.ID, time.Now().Add(21*24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "returned item")

		// Rule 4: Due date cannot be too far in future
		borrowing2, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// Try to extend beyond 90 days
		farFuture := time.Now().Add(100 * 24 * time.Hour)
		err = borrowingService.ExtendDueDate(borrowing2.ID, farFuture)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "90 days")

		// Rule 5: Cannot borrow with overdue items
		// Create overdue borrowing
		game2, err := gameService.AddGame("Overdue Test Game", "Game for overdue testing", "Test", "good")
		require.NoError(t, err)

		overdueBorrowing, err := borrowingService.BorrowGame(user.ID, game2.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// Manually update to be overdue
		overdueBorrowing.DueDate = time.Now().Add(-24 * time.Hour)
		overdueBorrowing.IsOverdue = true
		err = borrowingRepo.Update(overdueBorrowing)
		require.NoError(t, err)

		// Try to borrow another game (should fail)
		game3, err := gameService.AddGame("Another Test Game", "Another game", "Test", "good")
		require.NoError(t, err)

		_, err = borrowingService.BorrowGame(user.ID, game3.ID, time.Now().Add(14*24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "overdue")
	})
}

// TestDataValidationIntegration tests data validation across the system
func TestDataValidationIntegration(t *testing.T) {
	// Setup test database
	db, err := database.InitializeForTesting()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewSQLiteUserRepository(db)
	gameRepo := repositories.NewSQLiteGameRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)
	alertRepo := repositories.NewSQLiteAlertRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo, borrowingRepo)
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)
	alertService := services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	t.Run("User Validation", func(t *testing.T) {
		// Invalid email format
		_, err := userService.RegisterUser("Test User", "invalid-email")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Empty name
		_, err = userService.RegisterUser("", "valid@email.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Name too short
		_, err = userService.RegisterUser("A", "valid@email.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Valid user
		user, err := userService.RegisterUser("Valid User", "valid@email.com")
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("Game Validation", func(t *testing.T) {
		// Empty name
		_, err := gameService.AddGame("", "Description", "Category", "good")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Invalid condition
		_, err = gameService.AddGame("Test Game", "Description", "Category", "invalid-condition")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Name too short
		_, err = gameService.AddGame("A", "Description", "Category", "good")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Valid game
		game, err := gameService.AddGame("Valid Game", "Valid description", "Valid Category", "good")
		require.NoError(t, err)
		assert.NotZero(t, game.ID)
	})

	t.Run("Borrowing Validation", func(t *testing.T) {
		// Create valid user and game first
		user, err := userService.RegisterUser("Borrowing User", "borrowing@test.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Borrowing Game", "Game for borrowing", "Test", "good")
		require.NoError(t, err)

		// Invalid user ID
		_, err = borrowingService.BorrowGame(0, game.ID, time.Now().Add(14*24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")

		// Invalid game ID
		_, err = borrowingService.BorrowGame(user.ID, 0, time.Now().Add(14*24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid game ID")

		// Due date in the past
		_, err = borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(-24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Valid borrowing
		borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		assert.NotZero(t, borrowing.ID)
	})

	t.Run("Alert Validation", func(t *testing.T) {
		// Create valid user and game first
		user, err := userService.RegisterUser("Alert User", "alert@test.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Alert Game", "Game for alerts", "Test", "good")
		require.NoError(t, err)

		// Invalid user ID
		_, err = alertService.CreateCustomAlert(0, game.ID, "reminder", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")

		// Invalid game ID
		_, err = alertService.CreateCustomAlert(user.ID, 0, "reminder", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid game ID")

		// Invalid alert type
		_, err = alertService.CreateCustomAlert(user.ID, game.ID, "invalid-type", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Empty message
		_, err = alertService.CreateCustomAlert(user.ID, game.ID, "reminder", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		// Valid alert
		alert, err := alertService.CreateCustomAlert(user.ID, game.ID, "reminder", "Valid test message")
		require.NoError(t, err)
		assert.NotZero(t, alert.ID)
	})
}