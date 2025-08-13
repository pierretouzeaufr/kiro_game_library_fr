package services

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"fmt"
	"strings"
	"time"
)

// GameService handles game-related business logic
type GameService struct {
	gameRepo      repositories.GameRepository
	borrowingRepo repositories.BorrowingRepository
}

// NewGameService creates a new GameService instance
func NewGameService(gameRepo repositories.GameRepository, borrowingRepo repositories.BorrowingRepository) *GameService {
	return &GameService{
		gameRepo:      gameRepo,
		borrowingRepo: borrowingRepo,
	}
}

// AddGame creates a new game in the library
func (s *GameService) AddGame(name, description, category, condition string) (*models.Game, error) {
	// Create game model
	game := &models.Game{
		Name:        name,
		Description: description,
		Category:    category,
		EntryDate:   time.Now(),
		Condition:   condition,
		IsAvailable: true,
	}

	// Validate game data
	if err := models.ValidateGame(game); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create game in repository
	if err := s.gameRepo.Create(game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return game, nil
}

// GetGame retrieves a game by ID
func (s *GameService) GetGame(id int) (*models.Game, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid game ID: %d", id)
	}

	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return game, nil
}

// GetAllGames retrieves all games
func (s *GameService) GetAllGames() ([]*models.Game, error) {
	games, err := s.gameRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all games: %w", err)
	}

	return games, nil
}

// GetAvailableGames retrieves all available games
func (s *GameService) GetAvailableGames() ([]*models.Game, error) {
	games, err := s.gameRepo.GetAvailable()
	if err != nil {
		return nil, fmt.Errorf("failed to get available games: %w", err)
	}

	return games, nil
}

// SearchGames searches for games by query string
func (s *GameService) SearchGames(query string) ([]*models.Game, error) {
	if query == "" {
		return s.GetAllGames()
	}

	games, err := s.gameRepo.Search(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search games: %w", err)
	}

	return games, nil
}

// UpdateGame updates game information
func (s *GameService) UpdateGame(game *models.Game) error {
	if game == nil {
		return fmt.Errorf("game cannot be nil")
	}

	if game.ID <= 0 {
		return fmt.Errorf("invalid game ID: %d", game.ID)
	}

	// Validate game data
	if err := models.ValidateGame(game); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if game exists
	_, err := s.gameRepo.GetByID(game.ID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Update game in repository
	if err := s.gameRepo.Update(game); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

// SetGameAvailability updates the availability status of a game
func (s *GameService) SetGameAvailability(gameID int, isAvailable bool) error {
	if gameID <= 0 {
		return fmt.Errorf("invalid game ID: %d", gameID)
	}

	// Get the current game
	game, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Update availability status
	game.IsAvailable = isAvailable

	// Update game in repository
	if err := s.gameRepo.Update(game); err != nil {
		return fmt.Errorf("failed to update game availability: %w", err)
	}

	return nil
}

// GetGameBorrowingHistory retrieves the borrowing history for a specific game
func (s *GameService) GetGameBorrowingHistory(gameID int) ([]*models.Borrowing, error) {
	if gameID <= 0 {
		return nil, fmt.Errorf("invalid game ID: %d", gameID)
	}

	// Verify game exists
	_, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	borrowings, err := s.borrowingRepo.GetByGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game borrowing history: %w", err)
	}

	return borrowings, nil
}

// IsGameAvailable checks if a game is currently available for borrowing
func (s *GameService) IsGameAvailable(gameID int) (bool, error) {
	if gameID <= 0 {
		return false, fmt.Errorf("invalid game ID: %d", gameID)
	}

	game, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return false, fmt.Errorf("game not found: %w", err)
	}

	return game.IsAvailable, nil
}

// GetCurrentBorrower returns the current borrower of a game (if any)
func (s *GameService) GetCurrentBorrower(gameID int) (*models.Borrowing, error) {
	if gameID <= 0 {
		return nil, fmt.Errorf("invalid game ID: %d", gameID)
	}

	// Verify game exists
	_, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// Get all borrowings for this game
	borrowings, err := s.borrowingRepo.GetByGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game borrowings: %w", err)
	}

	// Find the active borrowing (not returned)
	for _, borrowing := range borrowings {
		if borrowing.ReturnedAt == nil {
			return borrowing, nil
		}
	}

	return nil, nil // No current borrower
}

// DeleteGame removes a game from the library
func (s *GameService) DeleteGame(gameID int) error {
	if gameID <= 0 {
		return fmt.Errorf("invalid game ID: %d", gameID)
	}

	// Check if game exists
	game, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Check if game has any active borrowings
	currentBorrower, err := s.GetCurrentBorrower(gameID)
	if err != nil {
		return fmt.Errorf("failed to check current borrower: %w", err)
	}

	if currentBorrower != nil {
		return fmt.Errorf("cannot delete game: currently borrowed by user %d", currentBorrower.UserID)
	}

	// Check if game has any borrowing history
	borrowingHistory, err := s.GetGameBorrowingHistory(gameID)
	if err != nil {
		return fmt.Errorf("failed to check borrowing history: %w", err)
	}

	if len(borrowingHistory) > 0 {
		return fmt.Errorf("cannot delete game '%s': it has borrowing history (%d records). Games with borrowing history cannot be deleted to maintain data integrity", game.Name, len(borrowingHistory))
	}

	// Delete game from repository
	if err := s.gameRepo.Delete(gameID); err != nil {
		// Check if it's a foreign key constraint error
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return fmt.Errorf("cannot delete game '%s': it has associated records (borrowings or alerts) that prevent deletion", game.Name)
		}
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
}