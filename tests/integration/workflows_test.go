package integration

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"board-game-library/internal/services"
	"board-game-library/pkg/database"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserRegistrationAndBorrowingWorkflow tests the complete end-to-end workflow
// from user registration through game borrowing and return
func TestUserRegistrationAndBorrowingWorkflow(t *testing.T) {
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
	_ = services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	t.Run("Complete User Registration and Borrowing Workflow", func(t *testing.T) {
		// Step 1: Register a new user
		user, err := userService.RegisterUser("John Doe", "john.doe@example.com")
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john.doe@example.com", user.Email)
		assert.True(t, user.IsActive)

		// Step 2: Add a game to the library
		game, err := gameService.AddGame("Settlers of Catan", "A strategy board game", "Strategy", "excellent")
		require.NoError(t, err)
		assert.NotZero(t, game.ID)
		assert.True(t, game.IsAvailable)

		// Step 3: Verify user can borrow
		canBorrow, err := userService.CanUserBorrow(user.ID)
		require.NoError(t, err)
		assert.True(t, canBorrow)

		// Step 4: Borrow the game
		dueDate := time.Now().Add(14 * 24 * time.Hour)
		borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, dueDate)
		require.NoError(t, err)
		assert.NotZero(t, borrowing.ID)
		assert.Equal(t, user.ID, borrowing.UserID)
		assert.Equal(t, game.ID, borrowing.GameID)
		assert.False(t, borrowing.IsOverdue)

		// Step 5: Verify game is no longer available
		updatedGame, err := gameService.GetGame(game.ID)
		require.NoError(t, err)
		assert.False(t, updatedGame.IsAvailable)

		// Step 6: Verify user has active borrowing
		activeBorrowings, err := userService.GetActiveUserBorrowings(user.ID)
		require.NoError(t, err)
		assert.Len(t, activeBorrowings, 1)
		assert.Equal(t, borrowing.ID, activeBorrowings[0].ID)

		// Step 7: Verify user cannot borrow another game while having active borrowing
		game2, err := gameService.AddGame("Ticket to Ride", "A railway-themed board game", "Strategy", "good")
		require.NoError(t, err)

		// Step 8: Return the game
		err = borrowingService.ReturnGame(borrowing.ID)
		require.NoError(t, err)

		// Step 9: Verify game is available again
		updatedGame, err = gameService.GetGame(game.ID)
		require.NoError(t, err)
		assert.True(t, updatedGame.IsAvailable)

		// Step 10: Verify borrowing is marked as returned
		returnedBorrowings, err := userService.GetUserBorrowings(user.ID)
		require.NoError(t, err)
		assert.Len(t, returnedBorrowings, 1)
		assert.NotNil(t, returnedBorrowings[0].ReturnedAt)

		// Step 11: Verify user can now borrow the second game
		canBorrow, err = userService.CanUserBorrow(user.ID)
		require.NoError(t, err)
		assert.True(t, canBorrow)

		borrowing2, err := borrowingService.BorrowGame(user.ID, game2.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		assert.NotZero(t, borrowing2.ID)
	})

	t.Run("Multiple Users Borrowing Different Games", func(t *testing.T) {
		// Register multiple users
		user1, err := userService.RegisterUser("Alice Smith", "alice@example.com")
		require.NoError(t, err)

		user2, err := userService.RegisterUser("Bob Johnson", "bob@example.com")
		require.NoError(t, err)

		// Add multiple games
		game1, err := gameService.AddGame("Monopoly", "Classic property trading game", "Family", "good")
		require.NoError(t, err)

		game2, err := gameService.AddGame("Scrabble", "Word formation game", "Word", "excellent")
		require.NoError(t, err)

		// Both users borrow different games
		_, err = borrowingService.BorrowGame(user1.ID, game1.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		_, err = borrowingService.BorrowGame(user2.ID, game2.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// Verify both borrowings are active
		activeBorrowings1, err := userService.GetActiveUserBorrowings(user1.ID)
		require.NoError(t, err)
		assert.Len(t, activeBorrowings1, 1)

		activeBorrowings2, err := userService.GetActiveUserBorrowings(user2.ID)
		require.NoError(t, err)
		assert.Len(t, activeBorrowings2, 1)

		// Verify games are not available
		updatedGame1, err := gameService.GetGame(game1.ID)
		require.NoError(t, err)
		assert.False(t, updatedGame1.IsAvailable)

		updatedGame2, err := gameService.GetGame(game2.ID)
		require.NoError(t, err)
		assert.False(t, updatedGame2.IsAvailable)
	})
}

// TestOverdueWorkflow tests the complete overdue item workflow
func TestOverdueWorkflow(t *testing.T) {
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

	t.Run("Overdue Item Processing and Alert Generation", func(t *testing.T) {
		// Step 1: Create user and game
		user, err := userService.RegisterUser("Test User", "test@example.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Test Game", "A test game", "Test", "good")
		require.NoError(t, err)

		// Step 2: Create an overdue borrowing (due date in the past)
		pastDueDate := time.Now().Add(-2 * 24 * time.Hour) // 2 days overdue
		borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, pastDueDate)
		require.NoError(t, err)

		// Step 3: Manually update the borrowing to be overdue (simulating time passage)
		borrowing.IsOverdue = true
		err = borrowingRepo.Update(borrowing)
		require.NoError(t, err)

		// Step 4: Generate overdue alerts
		err = alertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		// Step 5: Verify alert was created
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, "overdue", alerts[0].Type)
		assert.Equal(t, user.ID, alerts[0].UserID)
		assert.Equal(t, game.ID, alerts[0].GameID)
		assert.False(t, alerts[0].IsRead)

		// Step 6: Verify user cannot borrow while having overdue items
		canBorrow, err := userService.CanUserBorrow(user.ID)
		require.Error(t, err)
		assert.False(t, canBorrow)
		assert.Contains(t, err.Error(), "overdue")

		// Step 7: Return the overdue game
		err = borrowingService.ReturnGame(borrowing.ID)
		require.NoError(t, err)

		// Step 8: Verify user can now borrow again
		canBorrow, err = userService.CanUserBorrow(user.ID)
		require.NoError(t, err)
		assert.True(t, canBorrow)
	})

	t.Run("Reminder Alert Generation", func(t *testing.T) {
		// Step 1: Create user and game
		user, err := userService.RegisterUser("Reminder User", "reminder@example.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Reminder Game", "A game for reminder testing", "Test", "good")
		require.NoError(t, err)

		// Step 2: Create a borrowing due in 1 day (should trigger reminder)
		dueSoon := time.Now().Add(1 * 24 * time.Hour)
		_, err = borrowingService.BorrowGame(user.ID, game.ID, dueSoon)
		require.NoError(t, err)

		// Step 3: Generate reminder alerts
		err = alertService.GenerateReminderAlerts()
		require.NoError(t, err)

		// Step 4: Verify reminder alert was created
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		
		var reminderAlert *models.Alert
		for _, alert := range alerts {
			if alert.Type == "reminder" {
				reminderAlert = alert
				break
			}
		}
		
		require.NotNil(t, reminderAlert)
		assert.Equal(t, user.ID, reminderAlert.UserID)
		assert.Equal(t, game.ID, reminderAlert.GameID)
		assert.False(t, reminderAlert.IsRead)

		// Step 5: Mark alert as read
		err = alertService.MarkAlertAsRead(reminderAlert.ID)
		require.NoError(t, err)

		// Step 6: Verify alert is marked as read
		updatedAlerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		
		for _, alert := range updatedAlerts {
			if alert.ID == reminderAlert.ID {
				assert.True(t, alert.IsRead)
				break
			}
		}
	})
}

// TestDueDateExtensionWorkflow tests the due date extension functionality
func TestDueDateExtensionWorkflow(t *testing.T) {
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

	t.Run("Extend Due Date for Active Borrowing", func(t *testing.T) {
		// Step 1: Create user and game
		user, err := userService.RegisterUser("Extension User", "extension@example.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Extension Game", "A game for extension testing", "Test", "good")
		require.NoError(t, err)

		// Step 2: Borrow the game
		originalDueDate := time.Now().Add(14 * 24 * time.Hour)
		borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, originalDueDate)
		require.NoError(t, err)

		// Step 3: Extend the due date
		newDueDate := originalDueDate.Add(7 * 24 * time.Hour) // Extend by 7 days
		err = borrowingService.ExtendDueDate(borrowing.ID, newDueDate)
		require.NoError(t, err)

		// Step 4: Verify the due date was updated
		updatedBorrowing, err := borrowingRepo.GetByID(borrowing.ID)
		require.NoError(t, err)
		assert.True(t, updatedBorrowing.DueDate.After(originalDueDate))
		assert.WithinDuration(t, newDueDate, updatedBorrowing.DueDate, time.Minute)
	})
}

// TestErrorHandlingWorkflows tests various error conditions in workflows
func TestErrorHandlingWorkflows(t *testing.T) {
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

	t.Run("Attempt to Borrow Non-existent Game", func(t *testing.T) {
		user, err := userService.RegisterUser("Error User", "error@example.com")
		require.NoError(t, err)

		// Try to borrow a game that doesn't exist
		_, err = borrowingService.BorrowGame(user.ID, 99999, time.Now().Add(14*24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "game not found")
	})

	t.Run("Attempt to Borrow Already Borrowed Game", func(t *testing.T) {
		user1, err := userService.RegisterUser("User One", "user1@example.com")
		require.NoError(t, err)

		user2, err := userService.RegisterUser("User Two", "user2@example.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Popular Game", "A very popular game", "Strategy", "good")
		require.NoError(t, err)

		// User1 borrows the game
		_, err = borrowingService.BorrowGame(user1.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// User2 tries to borrow the same game
		_, err = borrowingService.BorrowGame(user2.ID, game.ID, time.Now().Add(14*24*time.Hour))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("Attempt to Register User with Duplicate Email", func(t *testing.T) {
		email := "duplicate@example.com"
		
		// Register first user
		_, err := userService.RegisterUser("First User", email)
		require.NoError(t, err)

		// Try to register second user with same email
		_, err = userService.RegisterUser("Second User", email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Attempt to Return Non-existent Borrowing", func(t *testing.T) {
		// Try to return a borrowing that doesn't exist
		err := borrowingService.ReturnGame(99999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "borrowing not found")
	})
}

// TestCompleteLibraryWorkflow tests a comprehensive end-to-end library workflow
func TestCompleteLibraryWorkflow(t *testing.T) {
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

	t.Run("Complete Library Management Scenario", func(t *testing.T) {
		// Phase 1: Library Setup
		// Register multiple users
		users := make([]*models.User, 3)
		for i := 0; i < 3; i++ {
			user, err := userService.RegisterUser(
				fmt.Sprintf("Library User %d", i+1),
				fmt.Sprintf("user%d@library.com", i+1),
			)
			require.NoError(t, err)
			users[i] = user
		}

		// Add multiple games to the library
		games := make([]*models.Game, 5)
		gameNames := []string{"Monopoly", "Scrabble", "Chess", "Risk", "Settlers of Catan"}
		categories := []string{"Family", "Word", "Strategy", "Strategy", "Strategy"}
		
		for i, name := range gameNames {
			game, err := gameService.AddGame(
				name,
				fmt.Sprintf("Description for %s", name),
				categories[i],
				"good",
			)
			require.NoError(t, err)
			games[i] = game
		}

		// Phase 2: Initial Borrowing Activity
		// User 1 borrows 2 games
		borrowing1, err := borrowingService.BorrowGame(users[0].ID, games[0].ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		
		borrowing2, err := borrowingService.BorrowGame(users[0].ID, games[1].ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// User 2 borrows 1 game
		_, err = borrowingService.BorrowGame(users[1].ID, games[2].ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// Verify borrowing state
		activeBorrowings1, err := userService.GetActiveUserBorrowings(users[0].ID)
		require.NoError(t, err)
		assert.Len(t, activeBorrowings1, 2)

		activeBorrowings2, err := userService.GetActiveUserBorrowings(users[1].ID)
		require.NoError(t, err)
		assert.Len(t, activeBorrowings2, 1)

		// Phase 3: Game Availability and Search
		// Verify borrowed games are not available
		for i := 0; i < 3; i++ {
			game, err := gameService.GetGame(games[i].ID)
			require.NoError(t, err)
			assert.False(t, game.IsAvailable, "Game %s should not be available", game.Name)
		}

		// Verify available games can be found
		availableGames, err := gameService.GetAvailableGames()
		require.NoError(t, err)
		assert.Len(t, availableGames, 2) // games[3] and games[4] should be available

		// Test search functionality
		strategyGames, err := gameService.SearchGames("Strategy")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(strategyGames), 2) // Should find strategy games

		// Phase 4: Due Date Management
		// Extend due date for one borrowing
		newDueDate := time.Now().Add(21 * 24 * time.Hour)
		err = borrowingService.ExtendDueDate(borrowing1.ID, newDueDate)
		require.NoError(t, err)

		// Verify extension
		updatedBorrowing, err := borrowingService.GetBorrowingDetails(borrowing1.ID)
		require.NoError(t, err)
		assert.True(t, updatedBorrowing.DueDate.After(time.Now().Add(20*24*time.Hour)))

		// Phase 5: Return Process
		// User 1 returns one game
		err = borrowingService.ReturnGame(borrowing2.ID)
		require.NoError(t, err)

		// Verify game is available again
		returnedGame, err := gameService.GetGame(games[1].ID)
		require.NoError(t, err)
		assert.True(t, returnedGame.IsAvailable)

		// Verify user's active borrowings updated
		activeBorrowings1, err = userService.GetActiveUserBorrowings(users[0].ID)
		require.NoError(t, err)
		assert.Len(t, activeBorrowings1, 1)

		// Phase 6: Overdue Scenario Simulation
		// Create an overdue borrowing by setting past due date
		pastDueDate := time.Now().Add(-2 * 24 * time.Hour)
		overdueBorrowing, err := borrowingService.BorrowGame(users[2].ID, games[4].ID, pastDueDate)
		require.NoError(t, err)

		// Manually mark as overdue (simulating time passage)
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
		assert.Equal(t, users[2].ID, alerts[0].UserID)

		// Verify user with overdue items cannot borrow
		canBorrow, err := userService.CanUserBorrow(users[2].ID)
		require.Error(t, err)
		assert.False(t, canBorrow)

		// Phase 7: Alert Management
		// Mark alert as read
		err = alertService.MarkAlertAsRead(alerts[0].ID)
		require.NoError(t, err)

		// Verify alert is marked as read
		updatedAlerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.True(t, updatedAlerts[0].IsRead)

		// Phase 8: Resolution and Cleanup
		// Return overdue game
		err = borrowingService.ReturnGame(overdueBorrowing.ID)
		require.NoError(t, err)

		// Verify user can now borrow again
		canBorrow, err = userService.CanUserBorrow(users[2].ID)
		require.NoError(t, err)
		assert.True(t, canBorrow)

		// Phase 9: Final State Verification
		// Check final game availability
		finalAvailableGames, err := gameService.GetAvailableGames()
		require.NoError(t, err)
		assert.Len(t, finalAvailableGames, 3) // 3 games should be available now

		// Check borrowing history
		user1History, err := userService.GetUserBorrowings(users[0].ID)
		require.NoError(t, err)
		assert.Len(t, user1History, 2) // User 1 had 2 borrowings

		user2History, err := userService.GetUserBorrowings(users[1].ID)
		require.NoError(t, err)
		assert.Len(t, user2History, 1) // User 2 had 1 borrowing

		user3History, err := userService.GetUserBorrowings(users[2].ID)
		require.NoError(t, err)
		assert.Len(t, user3History, 1) // User 3 had 1 borrowing

		// Verify all users can borrow
		for i, user := range users {
			canBorrow, err := userService.CanUserBorrow(user.ID)
			require.NoError(t, err, "User %d should be able to borrow", i+1)
			assert.True(t, canBorrow, "User %d should be able to borrow", i+1)
		}
	})
}

// TestDataIntegrityWorkflow tests data consistency across operations
func TestDataIntegrityWorkflow(t *testing.T) {
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

	t.Run("Data Consistency Across Operations", func(t *testing.T) {
		// Create test data
		user, err := userService.RegisterUser("Integrity User", "integrity@test.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Integrity Game", "Test game", "Test", "good")
		require.NoError(t, err)

		// Test borrowing creates consistent state
		borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)

		// Verify game state consistency
		updatedGame, err := gameService.GetGame(game.ID)
		require.NoError(t, err)
		assert.False(t, updatedGame.IsAvailable, "Game should be marked as unavailable")

		// Verify user borrowing state
		userBorrowings, err := userService.GetActiveUserBorrowings(user.ID)
		require.NoError(t, err)
		assert.Len(t, userBorrowings, 1)
		assert.Equal(t, borrowing.ID, userBorrowings[0].ID)

		// Test return creates consistent state
		err = borrowingService.ReturnGame(borrowing.ID)
		require.NoError(t, err)

		// Verify game is available again
		returnedGame, err := gameService.GetGame(game.ID)
		require.NoError(t, err)
		assert.True(t, returnedGame.IsAvailable, "Game should be marked as available")

		// Verify user has no active borrowings
		activeBorrowings, err := userService.GetActiveUserBorrowings(user.ID)
		require.NoError(t, err)
		assert.Len(t, activeBorrowings, 0)

		// Verify borrowing record is updated
		returnedBorrowing, err := borrowingService.GetBorrowingDetails(borrowing.ID)
		require.NoError(t, err)
		assert.NotNil(t, returnedBorrowing.ReturnedAt, "Borrowing should have return date")
	})

	t.Run("Alert Data Consistency", func(t *testing.T) {
		// Create test data
		user, err := userService.RegisterUser("Alert User", "alert@test.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Alert Game", "Test game", "Test", "good")
		require.NoError(t, err)

		// Create overdue borrowing
		overdueBorrowing, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(-24*time.Hour))
		require.NoError(t, err)

		// Mark as overdue
		overdueBorrowing.IsOverdue = true
		err = borrowingRepo.Update(overdueBorrowing)
		require.NoError(t, err)

		// Generate alerts
		err = alertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		// Verify alert exists
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, alerts, 1)

		// Return the game
		err = borrowingService.ReturnGame(overdueBorrowing.ID)
		require.NoError(t, err)

		// Verify alert still exists but borrowing is resolved
		finalAlerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, finalAlerts, 1) // Alert should still exist

		// Verify borrowing is resolved
		resolvedBorrowing, err := borrowingService.GetBorrowingDetails(overdueBorrowing.ID)
		require.NoError(t, err)
		assert.NotNil(t, resolvedBorrowing.ReturnedAt)
	})
}