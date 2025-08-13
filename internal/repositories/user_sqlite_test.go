package repositories

import (
	"board-game-library/internal/models"
	"board-game-library/pkg/database"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *database.DB {
	db, err := database.InitializeForTesting()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return db
}

func TestSQLiteUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	
	user := &models.User{
		Name:         "John Doe",
		Email:        "john@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set after creation")
	}
}

func TestSQLiteUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	
	// Create a user first
	user := &models.User{
		Name:         "Jane Doe",
		Email:        "jane@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}
	
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Retrieve the user
	retrieved, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrieved.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, retrieved.Name)
	}
	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}
}

func TestSQLiteUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	
	_, err := repo.GetByID(999)
	if err == nil {
		t.Error("Expected error for non-existent user, but got none")
	}
}

func TestSQLiteUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	
	// Create a user first
	user := &models.User{
		Name:         "Bob Smith",
		Email:        "bob@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}
	
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Retrieve the user by email
	retrieved, err := repo.GetByEmail(user.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if retrieved.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, retrieved.Name)
	}
	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, retrieved.ID)
	}
}

func TestSQLiteUserRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	
	// Create multiple users
	users := []*models.User{
		{Name: "Alice", Email: "alice@example.com", RegisteredAt: time.Now(), IsActive: true},
		{Name: "Bob", Email: "bob@example.com", RegisteredAt: time.Now(), IsActive: true},
		{Name: "Charlie", Email: "charlie@example.com", RegisteredAt: time.Now(), IsActive: false},
	}

	for _, user := range users {
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Retrieve all users
	retrieved, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(retrieved) != len(users) {
		t.Errorf("Expected %d users, got %d", len(users), len(retrieved))
	}
}

func TestSQLiteUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	
	// Create a user first
	user := &models.User{
		Name:         "Original Name",
		Email:        "original@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}
	
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update the user
	user.Name = "Updated Name"
	user.Email = "updated@example.com"
	user.IsActive = false

	err = repo.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Retrieve and verify the update
	retrieved, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.Email != "updated@example.com" {
		t.Errorf("Expected email 'updated@example.com', got %s", retrieved.Email)
	}
	if retrieved.IsActive != false {
		t.Errorf("Expected IsActive false, got %t", retrieved.IsActive)
	}
}

func TestSQLiteUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	
	// Create a user first
	user := &models.User{
		Name:         "To Delete",
		Email:        "delete@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}
	
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete the user
	err = repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify the user is deleted
	_, err = repo.GetByID(user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user, but got none")
	}
}

func TestSQLiteUserRepository_GetBorrowingHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	gameRepo := NewSQLiteGameRepository(db)
	borrowingRepo := NewSQLiteBorrowingRepository(db)
	
	// Create a user and game
	user := &models.User{
		Name:         "Test User",
		Email:        "test@example.com",
		RegisteredAt: time.Now(),
		IsActive:     true,
	}
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
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
		t.Fatalf("Failed to create game: %v", err)
	}

	// Create a borrowing
	borrowing := &models.Borrowing{
		UserID:     user.ID,
		GameID:     game.ID,
		BorrowedAt: time.Now(),
		DueDate:    time.Now().Add(14 * 24 * time.Hour),
		IsOverdue:  false,
	}
	err = borrowingRepo.Create(borrowing)
	if err != nil {
		t.Fatalf("Failed to create borrowing: %v", err)
	}

	// Get borrowing history
	history, err := userRepo.GetBorrowingHistory(user.ID)
	if err != nil {
		t.Fatalf("Failed to get borrowing history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 borrowing in history, got %d", len(history))
	}

	if history[0].GameID != game.ID {
		t.Errorf("Expected game ID %d, got %d", game.ID, history[0].GameID)
	}
}