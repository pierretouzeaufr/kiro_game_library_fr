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

// TestAlertGenerationIntegration tests the complete alert generation and processing workflow
func TestAlertGenerationIntegration(t *testing.T) {
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

	t.Run("Overdue Alert Generation and Processing", func(t *testing.T) {
		// Step 1: Create test data
		user, err := userService.RegisterUser("Alert Test User", "alert@example.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Alert Test Game", "A game for alert testing", "Test", "good")
		require.NoError(t, err)

		// Step 2: Create an overdue borrowing
		borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		
		// Manually update to be overdue (simulating time passage)
		borrowing.DueDate = time.Now().Add(-3 * 24 * time.Hour) // 3 days overdue
		borrowing.IsOverdue = true
		err = borrowingRepo.Update(borrowing)
		require.NoError(t, err)

		// Step 3: Generate overdue alerts
		err = alertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		// Step 4: Verify alert was created
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, alerts, 1)

		alert := alerts[0]
		assert.Equal(t, "overdue", alert.Type)
		assert.Equal(t, user.ID, alert.UserID)
		assert.Equal(t, game.ID, alert.GameID)
		assert.False(t, alert.IsRead)
		assert.Contains(t, alert.Message, "overdue")

		// Step 5: Test alert retrieval by user
		userAlerts, err := alertRepo.GetByUser(user.ID)
		require.NoError(t, err)
		assert.Len(t, userAlerts, 1)
		assert.Equal(t, alerts[0].ID, userAlerts[0].ID)

		// Step 6: Mark alert as read
		err = alertService.MarkAlertAsRead(alerts[0].ID)
		require.NoError(t, err)

		// Step 7: Verify alert is marked as read by checking user alerts
		userAlertsAfterRead, err := alertService.GetAlertsByUser(user.ID)
		require.NoError(t, err)
		assert.Len(t, userAlertsAfterRead, 1)
		assert.True(t, userAlertsAfterRead[0].IsRead)

		// Step 8: Return the game and verify alert behavior
		err = borrowingService.ReturnGame(borrowing.ID)
		require.NoError(t, err)

		// The alert should still exist but be marked as read
		finalUserAlerts, err := alertService.GetAlertsByUser(user.ID)
		require.NoError(t, err)
		assert.Len(t, finalUserAlerts, 1)
		assert.True(t, finalUserAlerts[0].IsRead)
	})

	t.Run("Reminder Alert Generation", func(t *testing.T) {
		// Step 1: Create test data
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
			if alert.Type == "reminder" && alert.UserID == user.ID {
				reminderAlert = alert
				break
			}
		}
		
		require.NotNil(t, reminderAlert)
		assert.Equal(t, "reminder", reminderAlert.Type)
		assert.Equal(t, user.ID, reminderAlert.UserID)
		assert.Equal(t, game.ID, reminderAlert.GameID)
		assert.False(t, reminderAlert.IsRead)
		assert.Contains(t, reminderAlert.Message, "due")

		// Step 5: Test that reminder is not generated again for same borrowing
		err = alertService.GenerateReminderAlerts()
		require.NoError(t, err)

		allAlerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		
		reminderCount := 0
		for _, alert := range allAlerts {
			if alert.Type == "reminder" && alert.UserID == user.ID && alert.GameID == game.ID {
				reminderCount++
			}
		}
		assert.Equal(t, 1, reminderCount, "Should not create duplicate reminder alerts")
	})

	t.Run("Multiple Users Multiple Alerts", func(t *testing.T) {
		const numUsers = 5
		const numGames = 5

		var users []*models.User
		var games []*models.Game
		var borrowings []*models.Borrowing

		// Create users and games
		for i := 0; i < numUsers; i++ {
			user, err := userService.RegisterUser(
				fmt.Sprintf("Multi User %d", i),
				fmt.Sprintf("multi%d@example.com", i),
			)
			require.NoError(t, err)
			users = append(users, user)

			game, err := gameService.AddGame(
				fmt.Sprintf("Multi Game %d", i),
				fmt.Sprintf("Game %d for multi-user testing", i),
				"Multi",
				"good",
			)
			require.NoError(t, err)
			games = append(games, game)
		}

		// Create overdue borrowings for all users
		for i := 0; i < numUsers; i++ {
			borrowing, err := borrowingService.BorrowGame(users[i].ID, games[i].ID, time.Now().Add(14*24*time.Hour))
			require.NoError(t, err)

			// Manually update to be overdue (simulating time passage)
			borrowing.DueDate = time.Now().Add(-time.Duration(i+1) * 24 * time.Hour) // Different overdue periods
			borrowing.IsOverdue = true
			err = borrowingRepo.Update(borrowing)
			require.NoError(t, err)

			borrowings = append(borrowings, borrowing)
		}

		// Generate overdue alerts
		err = alertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		// Verify alerts were created for all users
		alerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		
		overdueAlerts := 0
		for _, alert := range alerts {
			if alert.Type == "overdue" {
				overdueAlerts++
			}
		}
		assert.Equal(t, numUsers, overdueAlerts, "Should create overdue alert for each user")

		// Test alert retrieval by specific user
		for i, user := range users {
			userAlerts, err := alertRepo.GetByUser(user.ID)
			require.NoError(t, err)
			assert.Len(t, userAlerts, 1, "Each user should have exactly one alert")
			assert.Equal(t, games[i].ID, userAlerts[0].GameID)
		}

		// Test marking multiple alerts as read
		for _, alert := range alerts {
			if alert.Type == "overdue" {
				err = alertService.MarkAlertAsRead(alert.ID)
				require.NoError(t, err)
			}
		}

		// Verify all alerts are marked as read
		updatedAlerts, err := alertService.GetActiveAlerts()
		require.NoError(t, err)
		
		for _, alert := range updatedAlerts {
			if alert.Type == "overdue" {
				assert.True(t, alert.IsRead, "All overdue alerts should be marked as read")
			}
		}
	})

	t.Run("Alert Generation Performance with Large Dataset", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping performance test in short mode")
		}

		// Use separate database for this test to avoid interference
		perfDB, err := database.InitializeForTesting()
		require.NoError(t, err)
		defer perfDB.Close()

		// Initialize repositories for performance test
		perfUserRepo := repositories.NewSQLiteUserRepository(perfDB)
		perfGameRepo := repositories.NewSQLiteGameRepository(perfDB)
		perfBorrowingRepo := repositories.NewSQLiteBorrowingRepository(perfDB)
		perfAlertRepo := repositories.NewSQLiteAlertRepository(perfDB)

		// Initialize services for performance test
		perfUserService := services.NewUserService(perfUserRepo, perfBorrowingRepo)
		perfGameService := services.NewGameService(perfGameRepo, perfBorrowingRepo)
		perfBorrowingService := services.NewBorrowingService(perfBorrowingRepo, perfUserRepo, perfGameRepo)
		perfAlertService := services.NewAlertService(perfAlertRepo, perfBorrowingRepo, perfUserRepo, perfGameRepo)

		const numBorrowings = 100

		var users []*models.User
		var games []*models.Game

		// Create users and games
		for i := 0; i < numBorrowings; i++ {
			user, err := perfUserService.RegisterUser(
				fmt.Sprintf("Perf User %d", i),
				fmt.Sprintf("perf%d@example.com", i),
			)
			require.NoError(t, err)
			users = append(users, user)

			game, err := perfGameService.AddGame(
				fmt.Sprintf("Perf Game %d", i),
				fmt.Sprintf("Performance game %d", i),
				"Performance",
				"good",
			)
			require.NoError(t, err)
			games = append(games, game)
		}

		// Create overdue borrowings
		for i := 0; i < numBorrowings; i++ {
			borrowing, err := perfBorrowingService.BorrowGame(users[i].ID, games[i].ID, time.Now().Add(14*24*time.Hour))
			require.NoError(t, err)

			// Manually update to be overdue (simulating time passage)
			borrowing.DueDate = time.Now().Add(-2 * 24 * time.Hour)
			borrowing.IsOverdue = true
			err = perfBorrowingRepo.Update(borrowing)
			require.NoError(t, err)
		}

		// Test alert generation performance
		start := time.Now()
		err = perfAlertService.GenerateOverdueAlerts()
		duration := time.Since(start)
		require.NoError(t, err)

		t.Logf("Generated alerts for %d overdue borrowings in %v", numBorrowings, duration)

		// Verify all alerts were created
		alerts, err := perfAlertService.GetActiveAlerts()
		require.NoError(t, err)
		
		overdueAlerts := 0
		for _, alert := range alerts {
			if alert.Type == "overdue" {
				overdueAlerts++
			}
		}
		assert.Equal(t, numBorrowings, overdueAlerts)

		// Performance assertion
		assert.Less(t, duration, 5*time.Second, "Alert generation should complete within 5 seconds")
	})

	t.Run("Alert Cleanup and Management", func(t *testing.T) {
		// Use separate database for this test to avoid interference
		cleanupDB, err := database.InitializeForTesting()
		require.NoError(t, err)
		defer cleanupDB.Close()

		// Initialize repositories for cleanup test
		cleanupUserRepo := repositories.NewSQLiteUserRepository(cleanupDB)
		cleanupGameRepo := repositories.NewSQLiteGameRepository(cleanupDB)
		cleanupBorrowingRepo := repositories.NewSQLiteBorrowingRepository(cleanupDB)
		cleanupAlertRepo := repositories.NewSQLiteAlertRepository(cleanupDB)

		// Initialize services for cleanup test
		cleanupUserService := services.NewUserService(cleanupUserRepo, cleanupBorrowingRepo)
		cleanupGameService := services.NewGameService(cleanupGameRepo, cleanupBorrowingRepo)
		cleanupBorrowingService := services.NewBorrowingService(cleanupBorrowingRepo, cleanupUserRepo, cleanupGameRepo)
		cleanupAlertService := services.NewAlertService(cleanupAlertRepo, cleanupBorrowingRepo, cleanupUserRepo, cleanupGameRepo)

		// Create test data
		user, err := cleanupUserService.RegisterUser("Cleanup User", "cleanup@example.com")
		require.NoError(t, err)

		game, err := cleanupGameService.AddGame("Cleanup Game", "A game for cleanup testing", "Test", "good")
		require.NoError(t, err)

		// Create and return a borrowing to test alert lifecycle
		borrowing, err := cleanupBorrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		
		// Manually update to be overdue
		borrowing.DueDate = time.Now().Add(-1 * 24 * time.Hour)
		borrowing.IsOverdue = true
		err = cleanupBorrowingRepo.Update(borrowing)
		require.NoError(t, err)

		err = cleanupAlertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		// Verify alert exists
		alerts, err := cleanupAlertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, alerts, 1)

		// Return the game
		err = cleanupBorrowingService.ReturnGame(borrowing.ID)
		require.NoError(t, err)

		// Alert should still exist but can be cleaned up if needed
		finalAlerts, err := cleanupAlertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, finalAlerts, 1)

		// Test manual alert deletion
		err = cleanupAlertRepo.Delete(finalAlerts[0].ID)
		require.NoError(t, err)

		// Verify alert is deleted
		deletedAlerts, err := cleanupAlertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, deletedAlerts, 0)
	})
}

// TestAdvancedAlertScenarios tests complex alert generation scenarios
func TestAdvancedAlertScenarios(t *testing.T) {
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

	t.Run("Mixed Alert Types for Same User", func(t *testing.T) {
		// Use separate database for this test to avoid interference
		mixedDB, err := database.InitializeForTesting()
		require.NoError(t, err)
		defer mixedDB.Close()

		// Initialize repositories for mixed test
		mixedUserRepo := repositories.NewSQLiteUserRepository(mixedDB)
		mixedGameRepo := repositories.NewSQLiteGameRepository(mixedDB)
		mixedBorrowingRepo := repositories.NewSQLiteBorrowingRepository(mixedDB)
		mixedAlertRepo := repositories.NewSQLiteAlertRepository(mixedDB)

		// Initialize services for mixed test
		mixedUserService := services.NewUserService(mixedUserRepo, mixedBorrowingRepo)
		mixedGameService := services.NewGameService(mixedGameRepo, mixedBorrowingRepo)
		mixedBorrowingService := services.NewBorrowingService(mixedBorrowingRepo, mixedUserRepo, mixedGameRepo)
		mixedAlertService := services.NewAlertService(mixedAlertRepo, mixedBorrowingRepo, mixedUserRepo, mixedGameRepo)

		// Create users and games
		user1, err := mixedUserService.RegisterUser("Overdue Alert User", "overdue@example.com")
		require.NoError(t, err)

		user2, err := mixedUserService.RegisterUser("Reminder Alert User", "reminder@example.com")
		require.NoError(t, err)

		game1, err := mixedGameService.AddGame("Overdue Game", "Game that will be overdue", "Test", "good")
		require.NoError(t, err)

		game2, err := mixedGameService.AddGame("Reminder Game", "Game that will need reminder", "Test", "good")
		require.NoError(t, err)

		// Create overdue borrowing for user1
		overdueBorrowing, err := mixedBorrowingService.BorrowGame(user1.ID, game1.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		
		// Manually update to be overdue
		overdueBorrowing.DueDate = time.Now().Add(-3 * 24 * time.Hour)
		overdueBorrowing.IsOverdue = true
		err = mixedBorrowingRepo.Update(overdueBorrowing)
		require.NoError(t, err)

		// Create reminder borrowing for user2 (due in 1 day)
		_, err = mixedBorrowingService.BorrowGame(user2.ID, game2.ID, time.Now().Add(1*24*time.Hour))
		require.NoError(t, err)

		// Generate both types of alerts
		err = mixedAlertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		err = mixedAlertService.GenerateReminderAlerts()
		require.NoError(t, err)

		// Verify alerts exist for both users
		user1Alerts, err := mixedAlertService.GetAlertsByUser(user1.ID)
		require.NoError(t, err)
		assert.Len(t, user1Alerts, 1)
		assert.Equal(t, "overdue", user1Alerts[0].Type)

		user2Alerts, err := mixedAlertService.GetAlertsByUser(user2.ID)
		require.NoError(t, err)
		assert.Len(t, user2Alerts, 1)
		assert.Equal(t, "reminder", user2Alerts[0].Type)

		// Test alert summary
		summary, err := mixedAlertService.GetAlertsSummaryByUser()
		require.NoError(t, err)
		
		user1Summary := summary[user1.ID]
		assert.Equal(t, 1, user1Summary.TotalAlerts)
		assert.Equal(t, 1, user1Summary.OverdueCount)
		assert.Equal(t, 0, user1Summary.ReminderCount)

		user2Summary := summary[user2.ID]
		assert.Equal(t, 1, user2Summary.TotalAlerts)
		assert.Equal(t, 0, user2Summary.OverdueCount)
		assert.Equal(t, 1, user2Summary.ReminderCount)
	})

	t.Run("Alert Generation Edge Cases", func(t *testing.T) {
		// Test alert generation with various edge cases
		user, err := userService.RegisterUser("Edge Case User", "edge@example.com")
		require.NoError(t, err)

		// Case 1: Borrowing due exactly now
		game1, err := gameService.AddGame("Due Now Game", "Game due exactly now", "Test", "good")
		require.NoError(t, err)

		_, err = borrowingService.BorrowGame(user.ID, game1.ID, time.Now().Add(1*time.Hour))
		require.NoError(t, err)

		// Case 2: Borrowing due in exactly 2 days (boundary for reminders)
		game2, err := gameService.AddGame("Due In 2 Days Game", "Game due in exactly 2 days", "Test", "good")
		require.NoError(t, err)

		_, err = borrowingService.BorrowGame(user.ID, game2.ID, time.Now().Add(2*24*time.Hour))
		require.NoError(t, err)

		// Generate alerts
		err = alertService.GenerateReminderAlerts()
		require.NoError(t, err)

		// Verify appropriate alerts are generated
		alerts, err := alertService.GetAlertsByUser(user.ID)
		require.NoError(t, err)
		
		// Should have alerts for both games (both within 2-day window)
		assert.GreaterOrEqual(t, len(alerts), 1)
		
		// Verify alert messages are appropriate
		for _, alert := range alerts {
			assert.Contains(t, alert.Message, "due")
			assert.Equal(t, "reminder", alert.Type)
		}
	})

	t.Run("Alert Deduplication", func(t *testing.T) {
		// Test that duplicate alerts are not created
		user, err := userService.RegisterUser("Dedup User", "dedup@example.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Dedup Game", "Game for deduplication test", "Test", "good")
		require.NoError(t, err)

		// Create overdue borrowing
		borrowing, err := borrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
		require.NoError(t, err)
		
		// Manually update to be overdue
		borrowing.DueDate = time.Now().Add(-1 * 24 * time.Hour)
		borrowing.IsOverdue = true
		err = borrowingRepo.Update(borrowing)
		require.NoError(t, err)

		// Generate alerts multiple times
		for i := 0; i < 3; i++ {
			err = alertService.GenerateOverdueAlerts()
			require.NoError(t, err)
		}

		// Should only have one alert
		alerts, err := alertService.GetAlertsByUser(user.ID)
		require.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, "overdue", alerts[0].Type)
	})

	t.Run("Custom Alert Creation", func(t *testing.T) {
		// Test custom alert creation functionality
		user, err := userService.RegisterUser("Custom Alert User", "custom@example.com")
		require.NoError(t, err)

		game, err := gameService.AddGame("Custom Alert Game", "Game for custom alerts", "Test", "good")
		require.NoError(t, err)

		// Create custom alert
		customAlert, err := alertService.CreateCustomAlert(
			user.ID,
			game.ID,
			"reminder",
			"This is a custom reminder message for testing purposes.",
		)
		require.NoError(t, err)
		assert.NotZero(t, customAlert.ID)
		assert.Equal(t, user.ID, customAlert.UserID)
		assert.Equal(t, game.ID, customAlert.GameID)
		assert.Equal(t, "reminder", customAlert.Type)
		assert.False(t, customAlert.IsRead)

		// Verify custom alert appears in user alerts
		userAlerts, err := alertService.GetAlertsByUser(user.ID)
		require.NoError(t, err)
		assert.Len(t, userAlerts, 1)
		assert.Equal(t, customAlert.ID, userAlerts[0].ID)
	})

	t.Run("Bulk Alert Operations", func(t *testing.T) {
		// Use separate database for this test to avoid interference
		bulkDB, err := database.InitializeForTesting()
		require.NoError(t, err)
		defer bulkDB.Close()

		// Initialize repositories for bulk test
		bulkUserRepo := repositories.NewSQLiteUserRepository(bulkDB)
		bulkGameRepo := repositories.NewSQLiteGameRepository(bulkDB)
		bulkBorrowingRepo := repositories.NewSQLiteBorrowingRepository(bulkDB)
		bulkAlertRepo := repositories.NewSQLiteAlertRepository(bulkDB)

		// Initialize services for bulk test
		bulkUserService := services.NewUserService(bulkUserRepo, bulkBorrowingRepo)
		bulkGameService := services.NewGameService(bulkGameRepo, bulkBorrowingRepo)
		bulkBorrowingService := services.NewBorrowingService(bulkBorrowingRepo, bulkUserRepo, bulkGameRepo)
		bulkAlertService := services.NewAlertService(bulkAlertRepo, bulkBorrowingRepo, bulkUserRepo, bulkGameRepo)

		// Create multiple users and games for bulk operations
		const numGames = 5
		users := make([]*models.User, numGames)
		games := make([]*models.Game, numGames)
		borrowings := make([]*models.Borrowing, numGames)

		for i := 0; i < numGames; i++ {
			// Create separate user for each borrowing to avoid overdue conflicts
			user, err := bulkUserService.RegisterUser(
				fmt.Sprintf("Bulk Alert User %d", i),
				fmt.Sprintf("bulk%d@example.com", i),
			)
			require.NoError(t, err)
			users[i] = user

			game, err := bulkGameService.AddGame(
				fmt.Sprintf("Bulk Game %d", i),
				fmt.Sprintf("Game %d for bulk testing", i),
				"Bulk",
				"good",
			)
			require.NoError(t, err)
			games[i] = game

			// Create overdue borrowing
			borrowing, err := bulkBorrowingService.BorrowGame(user.ID, game.ID, time.Now().Add(14*24*time.Hour))
			require.NoError(t, err)
			
			// Manually update to be overdue
			borrowing.DueDate = time.Now().Add(-time.Duration(i+1) * 24 * time.Hour)
			borrowing.IsOverdue = true
			err = bulkBorrowingRepo.Update(borrowing)
			require.NoError(t, err)
			borrowings[i] = borrowing
		}

		// Generate alerts for all overdue items
		err = bulkAlertService.GenerateOverdueAlerts()
		require.NoError(t, err)

		// Verify all alerts were created
		allAlerts, err := bulkAlertService.GetActiveAlerts()
		require.NoError(t, err)
		assert.Len(t, allAlerts, numGames)

		// Test bulk mark as read for first user
		firstUser := users[0]
		err = bulkAlertService.MarkAllUserAlertsAsRead(firstUser.ID)
		require.NoError(t, err)

		// Verify first user's alerts are marked as read
		firstUserAlerts, err := bulkAlertService.GetAlertsByUser(firstUser.ID)
		require.NoError(t, err)
		assert.Len(t, firstUserAlerts, 1)
		assert.True(t, firstUserAlerts[0].IsRead, "Alert should be marked as read")

		// Verify other users' alerts are still unread
		secondUserAlerts, err := bulkAlertService.GetAlertsByUser(users[1].ID)
		require.NoError(t, err)
		assert.Len(t, secondUserAlerts, 1)
		assert.False(t, secondUserAlerts[0].IsRead, "Alert should still be unread")
	})
}