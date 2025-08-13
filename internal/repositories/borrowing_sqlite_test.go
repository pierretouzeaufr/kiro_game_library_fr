package repositories

import (
	"board-game-library/internal/models"
	"testing"
	"time"
)

func createTestUserAndGame(t *testing.T, userRepo UserRepository, gameRepo GameRepository) (*models.User, *models.Game) {
	user := &models.User{
		Name:         "Test User",
		Email:        "test@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	game := &models.Game{
		Name:        "Test Game",
		Description: "A test game",
		Category:    "Strategy",
		EntryDate:   time.Now(),
		Condition:   "good",
		IsAvailable: true,
	}
	err = gameRepo.Create(game)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	return user, game
}

func TestSQLiteBorrowingRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	borrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now(),
		DueDate:    time.Now().Add(14 * 24 * time.Hour),
		IsOverdue:  false,
	}

	err := borrowingRepo.Create(borrowing)
	if err != nil {
		t.Fatalf("Failed to create borrowing: %v", err)
	}

	if borrowing.ID == 0 {
		t.Error("Expected borrowing ID to be set after creation")
	}
}

func TestSQLiteBorrowingRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	borrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now(),
		DueDate:    time.Now().Add(14 * 24 * time.Hour),
		IsOverdue:  false,
	}

	err := borrowingRepo.Create(borrowing)
	if err != nil {
		t.Fatalf("Failed to create borrowing: %v", err)
	}

	// Retrieve the borrowing
	retrieved, err := borrowingRepo.GetByID(borrowing.ID)
	if err != nil {
		t.Fatalf("Failed to get borrowing by ID: %v", err)
	}

	if retrieved.UserID != borrowing.UserID {
		t.Errorf("Expected user ID %d, got %d", borrowing.UserID, retrieved.UserID)
	}
	if retrieved.GameID != borrowing.GameID {
		t.Errorf("Expected game ID %d, got %d", borrowing.GameID, retrieved.GameID)
	}
}

func TestSQLiteBorrowingRepository_GetActiveByUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	// Create an active borrowing
	activeBorrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now(),
		DueDate:    time.Now().Add(14 * 24 * time.Hour),
		IsOverdue:  false,
	}
	err := borrowingRepo.Create(activeBorrowing)
	if err != nil {
		t.Fatalf("Failed to create active borrowing: %v", err)
	}

	// Create a returned borrowing
	returnedTime := time.Now()
	returnedBorrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
		DueDate:    time.Now().Add(7 * 24 * time.Hour),
		ReturnedAt: &returnedTime,
		IsOverdue:  false,
	}
	err = borrowingRepo.Create(returnedBorrowing)
	if err != nil {
		t.Fatalf("Failed to create returned borrowing: %v", err)
	}

	// Get active borrowings
	active, err := borrowingRepo.GetActiveByUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get active borrowings: %v", err)
	}

	if len(active) != 1 {
		t.Errorf("Expected 1 active borrowing, got %d", len(active))
	}

	if active[0].ID != activeBorrowing.ID {
		t.Errorf("Expected active borrowing ID %d, got %d", activeBorrowing.ID, active[0].ID)
	}
}

func TestSQLiteBorrowingRepository_GetByGame(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	// Create multiple borrowings for the same game
	borrowings := []*models.Borrowing{
		{UserID: user.ID, GameID: game.ID, BorrowedAt: time.Now().Add(-30 * 24 * time.Hour), DueDate: time.Now().Add(-16 * 24 * time.Hour), IsOverdue: false},
		{UserID: user.ID, GameID: game.ID, BorrowedAt: time.Now().Add(-14 * 24 * time.Hour), DueDate: time.Now(), IsOverdue: false},
	}

	for _, borrowing := range borrowings {
		err := borrowingRepo.Create(borrowing)
		if err != nil {
			t.Fatalf("Failed to create borrowing: %v", err)
		}
	}

	// Get borrowings by game
	gameBorrowings, err := borrowingRepo.GetByGame(game.ID)
	if err != nil {
		t.Fatalf("Failed to get borrowings by game: %v", err)
	}

	if len(gameBorrowings) != 2 {
		t.Errorf("Expected 2 borrowings for game, got %d", len(gameBorrowings))
	}
}

func TestSQLiteBorrowingRepository_GetOverdue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	// Create an overdue borrowing (due date in the past, not returned)
	overdueBorrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now().Add(-20 * 24 * time.Hour),
		DueDate:    time.Now().Add(-5 * 24 * time.Hour), // 5 days overdue
		IsOverdue:  true,
	}
	err := borrowingRepo.Create(overdueBorrowing)
	if err != nil {
		t.Fatalf("Failed to create overdue borrowing: %v", err)
	}

	// Create a current borrowing (not overdue)
	currentBorrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now(),
		DueDate:    time.Now().Add(14 * 24 * time.Hour), // Due in 14 days
		IsOverdue:  false,
	}
	err = borrowingRepo.Create(currentBorrowing)
	if err != nil {
		t.Fatalf("Failed to create current borrowing: %v", err)
	}

	// Get overdue borrowings
	overdue, err := borrowingRepo.GetOverdue()
	if err != nil {
		t.Fatalf("Failed to get overdue borrowings: %v", err)
	}

	if len(overdue) != 1 {
		t.Errorf("Expected 1 overdue borrowing, got %d", len(overdue))
	}

	if overdue[0].ID != overdueBorrowing.ID {
		t.Errorf("Expected overdue borrowing ID %d, got %d", overdueBorrowing.ID, overdue[0].ID)
	}
}

func TestSQLiteBorrowingRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	borrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now(),
		DueDate:    time.Now().Add(14 * 24 * time.Hour),
		IsOverdue:  false,
	}

	err := borrowingRepo.Create(borrowing)
	if err != nil {
		t.Fatalf("Failed to create borrowing: %v", err)
	}

	// Update the borrowing
	borrowing.IsOverdue = true
	borrowing.DueDate = time.Now().Add(-1 * 24 * time.Hour) // Make it overdue

	err = borrowingRepo.Update(borrowing)
	if err != nil {
		t.Fatalf("Failed to update borrowing: %v", err)
	}

	// Retrieve and verify the update
	retrieved, err := borrowingRepo.GetByID(borrowing.ID)
	if err != nil {
		t.Fatalf("Failed to get updated borrowing: %v", err)
	}

	if !retrieved.IsOverdue {
		t.Error("Expected borrowing to be marked as overdue")
	}
}

func TestSQLiteBorrowingRepository_ReturnGame(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	borrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now(),
		DueDate:    time.Now().Add(14 * 24 * time.Hour),
		IsOverdue:  false,
	}

	err := borrowingRepo.Create(borrowing)
	if err != nil {
		t.Fatalf("Failed to create borrowing: %v", err)
	}

	// Return the game
	err = borrowingRepo.ReturnGame(borrowing.ID)
	if err != nil {
		t.Fatalf("Failed to return game: %v", err)
	}

	// Retrieve and verify the return
	retrieved, err := borrowingRepo.GetByID(borrowing.ID)
	if err != nil {
		t.Fatalf("Failed to get returned borrowing: %v", err)
	}

	if retrieved.ReturnedAt == nil {
		t.Error("Expected ReturnedAt to be set after returning game")
	}
}

func TestSQLiteBorrowingRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)

	user, game := createTestUserAndGame(t, userRepo, gameRepo)

	// Create multiple borrowings
	borrowings := []*models.Borrowing{
		{UserID: user.ID, GameID: game.ID, BorrowedAt: time.Now().Add(-30 * 24 * time.Hour), DueDate: time.Now().Add(-16 * 24 * time.Hour), IsOverdue: false},
		{UserID: user.ID, GameID: game.ID, BorrowedAt: time.Now().Add(-14 * 24 * time.Hour), DueDate: time.Now(), IsOverdue: false},
		{UserID: user.ID, GameID: game.ID, BorrowedAt: time.Now(), DueDate: time.Now().Add(14 * 24 * time.Hour), IsOverdue: false},
	}

	for _, borrowing := range borrowings {
		err := borrowingRepo.Create(borrowing)
		if err != nil {
			t.Fatalf("Failed to create borrowing: %v", err)
		}
	}

	// Get all borrowings
	all, err := borrowingRepo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all borrowings: %v", err)
	}

	if len(all) != len(borrowings) {
		t.Errorf("Expected %d borrowings, got %d", len(borrowings), len(all))
	}
}