package repositories

import (
	"board-game-library/internal/models"
	"testing"
	"time"
)

func TestSQLiteAlertRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	alertRepo := NewSQLiteAlertRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	alert := &models.Alert{
		UserID:    user.ID,
		GameID:    game.ID,
		Type:      "overdue",
		Message:   "Game is overdue",
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	err := alertRepo.Create(alert)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	if alert.ID == 0 {
		t.Error("Expected alert ID to be set after creation")
	}
}

func TestSQLiteAlertRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	alertRepo := NewSQLiteAlertRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	alert := &models.Alert{
		UserID:    user.ID,
		GameID:    game.ID,
		Type:      "reminder",
		Message:   "Game due soon",
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	err := alertRepo.Create(alert)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	// Retrieve the alert
	retrieved, err := alertRepo.GetByID(alert.ID)
	if err != nil {
		t.Fatalf("Failed to get alert by ID: %v", err)
	}

	if retrieved.UserID != alert.UserID {
		t.Errorf("Expected user ID %d, got %d", alert.UserID, retrieved.UserID)
	}
	if retrieved.GameID != alert.GameID {
		t.Errorf("Expected game ID %d, got %d", alert.GameID, retrieved.GameID)
	}
	if retrieved.Type != alert.Type {
		t.Errorf("Expected type %s, got %s", alert.Type, retrieved.Type)
	}
	if retrieved.Message != alert.Message {
		t.Errorf("Expected message %s, got %s", alert.Message, retrieved.Message)
	}
}

func TestSQLiteAlertRepository_GetUnread(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	alertRepo := NewSQLiteAlertRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	// Create unread alerts
	unreadAlerts := []*models.Alert{
		{UserID: user.ID, GameID: game.ID, Type: "overdue", Message: "Overdue alert", CreatedAt: time.Now(), IsRead: false},
		{UserID: user.ID, GameID: game.ID, Type: "reminder", Message: "Reminder alert", CreatedAt: time.Now(), IsRead: false},
	}

	// Create read alert
	readAlert := &models.Alert{
		UserID:    user.ID,
		GameID:    game.ID,
		Type:      "overdue",
		Message:   "Read alert",
		CreatedAt: time.Now(),
		IsRead:    true,
	}

	for _, alert := range unreadAlerts {
		err := alertRepo.Create(alert)
		if err != nil {
			t.Fatalf("Failed to create unread alert: %v", err)
		}
	}

	err := alertRepo.Create(readAlert)
	if err != nil {
		t.Fatalf("Failed to create read alert: %v", err)
	}

	// Get unread alerts
	unread, err := alertRepo.GetUnread()
	if err != nil {
		t.Fatalf("Failed to get unread alerts: %v", err)
	}

	if len(unread) != 2 {
		t.Errorf("Expected 2 unread alerts, got %d", len(unread))
	}

	// Verify all returned alerts are unread
	for _, alert := range unread {
		if alert.IsRead {
			t.Errorf("Expected all alerts to be unread, but found read alert: %s", alert.Message)
		}
	}
}

func TestSQLiteAlertRepository_GetByUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	alertRepo := NewSQLiteAlertRepository(db)

	user1, game := createTestUserAndGame(t, userRepo, gameRepo)
	
	// Create second user
	user2 := &models.User{
		Name:         "Second User",
		Email:        "second@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}
	err := userRepo.Create(user2)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	// Create alerts for different users
	user1Alerts := []*models.Alert{
		{UserID: user1.ID, GameID: game.ID, Type: "overdue", Message: "User1 alert 1", CreatedAt: time.Now(), IsRead: false},
		{UserID: user1.ID, GameID: game.ID, Type: "reminder", Message: "User1 alert 2", CreatedAt: time.Now(), IsRead: true},
	}

	user2Alert := &models.Alert{
		UserID:    user2.ID,
		GameID:    game.ID,
		Type:      "overdue",
		Message:   "User2 alert",
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	for _, alert := range user1Alerts {
		err := alertRepo.Create(alert)
		if err != nil {
			t.Fatalf("Failed to create user1 alert: %v", err)
		}
	}

	err = alertRepo.Create(user2Alert)
	if err != nil {
		t.Fatalf("Failed to create user2 alert: %v", err)
	}

	// Get alerts for user1
	user1AlertsRetrieved, err := alertRepo.GetByUser(user1.ID)
	if err != nil {
		t.Fatalf("Failed to get alerts by user: %v", err)
	}

	if len(user1AlertsRetrieved) != 2 {
		t.Errorf("Expected 2 alerts for user1, got %d", len(user1AlertsRetrieved))
	}

	// Verify all returned alerts belong to user1
	for _, alert := range user1AlertsRetrieved {
		if alert.UserID != user1.ID {
			t.Errorf("Expected all alerts to belong to user1 (%d), but found alert for user %d", user1.ID, alert.UserID)
		}
	}
}

func TestSQLiteAlertRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	alertRepo := NewSQLiteAlertRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	// Create multiple alerts
	alerts := []*models.Alert{
		{UserID: user.ID, GameID: game.ID, Type: "overdue", Message: "Alert 1", CreatedAt: time.Now().Add(-2 * time.Hour), IsRead: false},
		{UserID: user.ID, GameID: game.ID, Type: "reminder", Message: "Alert 2", CreatedAt: time.Now().Add(-1 * time.Hour), IsRead: true},
		{UserID: user.ID, GameID: game.ID, Type: "overdue", Message: "Alert 3", CreatedAt: time.Now(), IsRead: false},
	}

	for _, alert := range alerts {
		err := alertRepo.Create(alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}
	}

	// Get all alerts
	all, err := alertRepo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all alerts: %v", err)
	}

	if len(all) != len(alerts) {
		t.Errorf("Expected %d alerts, got %d", len(alerts), len(all))
	}
}

func TestSQLiteAlertRepository_MarkAsRead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	alertRepo := NewSQLiteAlertRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	alert := &models.Alert{
		UserID:    user.ID,
		GameID:    game.ID,
		Type:      "overdue",
		Message:   "Unread alert",
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	err := alertRepo.Create(alert)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	// Mark as read
	err = alertRepo.MarkAsRead(alert.ID)
	if err != nil {
		t.Fatalf("Failed to mark alert as read: %v", err)
	}

	// Retrieve and verify
	retrieved, err := alertRepo.GetByID(alert.ID)
	if err != nil {
		t.Fatalf("Failed to get alert after marking as read: %v", err)
	}

	if !retrieved.IsRead {
		t.Error("Expected alert to be marked as read")
	}
}

func TestSQLiteAlertRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	alertRepo := NewSQLiteAlertRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	alert := &models.Alert{
		UserID:    user.ID,
		GameID:    game.ID,
		Type:      "overdue",
		Message:   "To delete",
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	err := alertRepo.Create(alert)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	// Delete the alert
	err = alertRepo.Delete(alert.ID)
	if err != nil {
		t.Fatalf("Failed to delete alert: %v", err)
	}

	// Verify the alert is deleted
	_, err = alertRepo.GetByID(alert.ID)
	if err == nil {
		t.Error("Expected error when getting deleted alert, but got none")
	}
}