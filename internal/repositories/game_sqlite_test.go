package repositories

import (
	"board-game-library/internal/models"
	"testing"
	"time"
)

func TestSQLiteGameRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteGameRepository(db)
	
	game := &models.Game{
		Name:        "Monopoly",
		Description: "Classic board game",
		Category:    "Strategy",
		EntryDate:   time.Now(),
		Condition:   "good",
		IsAvailable: true,
	}

	err := repo.Create(game)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	if game.ID == 0 {
		t.Error("Expected game ID to be set after creation")
	}
}

func TestSQLiteGameRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteGameRepository(db)
	
	// Create a game first
	game := &models.Game{
		Name:        "Scrabble",
		Description: "Word game",
		Category:    "Word",
		EntryDate:   time.Now(),
		Condition:   "excellent",
		IsAvailable: true,
	}
	
	err := repo.Create(game)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Retrieve the game
	retrieved, err := repo.GetByID(game.ID)
	if err != nil {
		t.Fatalf("Failed to get game by ID: %v", err)
	}

	if retrieved.Name != game.Name {
		t.Errorf("Expected name %s, got %s", game.Name, retrieved.Name)
	}
	if retrieved.Description != game.Description {
		t.Errorf("Expected description %s, got %s", game.Description, retrieved.Description)
	}
	if retrieved.Category != game.Category {
		t.Errorf("Expected category %s, got %s", game.Category, retrieved.Category)
	}
}

func TestSQLiteGameRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteGameRepository(db)
	
	// Create multiple games
	games := []*models.Game{
		{Name: "Chess", Description: "Strategy game", Category: "Strategy", EntryDate: time.Now(), Condition: "good", IsAvailable: true},
		{Name: "Checkers", Description: "Classic game", Category: "Strategy", EntryDate: time.Now(), Condition: "fair", IsAvailable: false},
		{Name: "Risk", Description: "World domination", Category: "Strategy", EntryDate: time.Now(), Condition: "excellent", IsAvailable: true},
	}

	for _, game := range games {
		err := repo.Create(game)
		if err != nil {
			t.Fatalf("Failed to create game: %v", err)
		}
	}

	// Retrieve all games
	retrieved, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all games: %v", err)
	}

	if len(retrieved) != len(games) {
		t.Errorf("Expected %d games, got %d", len(games), len(retrieved))
	}
}

func TestSQLiteGameRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteGameRepository(db)
	
	// Create games with different names and descriptions
	games := []*models.Game{
		{Name: "Monopoly", Description: "Property trading game", Category: "Strategy", EntryDate: time.Now(), Condition: "good", IsAvailable: true},
		{Name: "Scrabble", Description: "Word building game", Category: "Word", EntryDate: time.Now(), Condition: "good", IsAvailable: true},
		{Name: "Chess", Description: "Strategic board game", Category: "Strategy", EntryDate: time.Now(), Condition: "good", IsAvailable: true},
	}

	for _, game := range games {
		err := repo.Create(game)
		if err != nil {
			t.Fatalf("Failed to create game: %v", err)
		}
	}

	// Search by name
	results, err := repo.Search("Monopoly")
	if err != nil {
		t.Fatalf("Failed to search games: %v", err)
	}
	if len(results) != 1 || results[0].Name != "Monopoly" {
		t.Errorf("Expected 1 result with name 'Monopoly', got %d results", len(results))
	}

	// Search by description
	results, err = repo.Search("game")
	if err != nil {
		t.Fatalf("Failed to search games: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results for 'game' search, got %d", len(results))
	}

	// Search by category
	results, err = repo.Search("Strategy")
	if err != nil {
		t.Fatalf("Failed to search games: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Strategy' search, got %d", len(results))
	}
}

func TestSQLiteGameRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteGameRepository(db)
	
	// Create a game first
	game := &models.Game{
		Name:        "Original Name",
		Description: "Original description",
		Category:    "Original",
		EntryDate:   time.Now(),
		Condition:   "good",
		IsAvailable: true,
	}
	
	err := repo.Create(game)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Update the game
	game.Name = "Updated Name"
	game.Description = "Updated description"
	game.Category = "Updated"
	game.Condition = "excellent"
	game.IsAvailable = false

	err = repo.Update(game)
	if err != nil {
		t.Fatalf("Failed to update game: %v", err)
	}

	// Retrieve and verify the update
	retrieved, err := repo.GetByID(game.ID)
	if err != nil {
		t.Fatalf("Failed to get updated game: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got %s", retrieved.Description)
	}
	if retrieved.IsAvailable != false {
		t.Errorf("Expected IsAvailable false, got %t", retrieved.IsAvailable)
	}
}

func TestSQLiteGameRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteGameRepository(db)
	
	// Create a game first
	game := &models.Game{
		Name:        "To Delete",
		Description: "Game to be deleted",
		Category:    "Test",
		EntryDate:   time.Now(),
		Condition:   "good",
		IsAvailable: true,
	}
	
	err := repo.Create(game)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Delete the game
	err = repo.Delete(game.ID)
	if err != nil {
		t.Fatalf("Failed to delete game: %v", err)
	}

	// Verify the game is deleted
	_, err = repo.GetByID(game.ID)
	if err == nil {
		t.Error("Expected error when getting deleted game, but got none")
	}
}

func TestSQLiteGameRepository_GetAvailable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteGameRepository(db)
	
	// Create games with different availability
	games := []*models.Game{
		{Name: "Available Game 1", Description: "Available", Category: "Test", EntryDate: time.Now(), Condition: "good", IsAvailable: true},
		{Name: "Unavailable Game", Description: "Not available", Category: "Test", EntryDate: time.Now(), Condition: "good", IsAvailable: false},
		{Name: "Available Game 2", Description: "Available", Category: "Test", EntryDate: time.Now(), Condition: "good", IsAvailable: true},
	}

	for _, game := range games {
		err := repo.Create(game)
		if err != nil {
			t.Fatalf("Failed to create game: %v", err)
		}
	}

	// Get available games
	available, err := repo.GetAvailable()
	if err != nil {
		t.Fatalf("Failed to get available games: %v", err)
	}

	if len(available) != 2 {
		t.Errorf("Expected 2 available games, got %d", len(available))
	}

	// Verify all returned games are available
	for _, game := range available {
		if !game.IsAvailable {
			t.Errorf("Expected all games to be available, but found unavailable game: %s", game.Name)
		}
	}
}