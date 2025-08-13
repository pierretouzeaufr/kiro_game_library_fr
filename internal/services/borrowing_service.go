package services

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"fmt"
	"time"
)

// BorrowingService handles borrowing-related business logic
type BorrowingService struct {
	borrowingRepo repositories.BorrowingRepository
	userRepo      repositories.UserRepository
	gameRepo      repositories.GameRepository
}

// NewBorrowingService creates a new BorrowingService instance
func NewBorrowingService(borrowingRepo repositories.BorrowingRepository, userRepo repositories.UserRepository, gameRepo repositories.GameRepository) *BorrowingService {
	return &BorrowingService{
		borrowingRepo: borrowingRepo,
		userRepo:      userRepo,
		gameRepo:      gameRepo,
	}
}

// BorrowGame creates a new borrowing record
func (s *BorrowingService) BorrowGame(userID, gameID int, dueDate time.Time) (*models.Borrowing, error) {
	// Validate input parameters
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}
	if gameID <= 0 {
		return nil, fmt.Errorf("invalid game ID: %d", gameID)
	}

	// Check if user exists and is active
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Check if user has any overdue items
	activeBorrowings, err := s.borrowingRepo.GetActiveByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user borrowings: %w", err)
	}
	for _, borrowing := range activeBorrowings {
		if borrowing.IsCurrentlyOverdue() {
			return nil, fmt.Errorf("user has overdue items and cannot borrow")
		}
	}

	// Check if game exists and is available
	game, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}
	if !game.IsAvailable {
		return nil, fmt.Errorf("game is not available for borrowing")
	}

	// Create borrowing record
	borrowing := &models.Borrowing{
		UserID:     userID,
		GameID:     gameID,
		BorrowedAt: time.Now(),
		DueDate:    dueDate,
		ReturnedAt: nil,
		IsOverdue:  false,
	}

	// Validate borrowing data
	if err := models.ValidateBorrowing(borrowing); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create borrowing in repository
	if err := s.borrowingRepo.Create(borrowing); err != nil {
		return nil, fmt.Errorf("failed to create borrowing: %w", err)
	}

	// Update game availability
	game.IsAvailable = false
	if err := s.gameRepo.Update(game); err != nil {
		return nil, fmt.Errorf("failed to update game availability: %w", err)
	}

	return borrowing, nil
}

// BorrowGameWithDefaultDueDate creates a borrowing with default 14-day due date
func (s *BorrowingService) BorrowGameWithDefaultDueDate(userID, gameID int) (*models.Borrowing, error) {
	dueDate := time.Now().Add(14 * 24 * time.Hour) // 14 days from now
	return s.BorrowGame(userID, gameID, dueDate)
}

// ReturnGame processes the return of a borrowed game
func (s *BorrowingService) ReturnGame(borrowingID int) error {
	if borrowingID <= 0 {
		return fmt.Errorf("invalid borrowing ID: %d", borrowingID)
	}

	// Get the borrowing record
	borrowing, err := s.borrowingRepo.GetByID(borrowingID)
	if err != nil {
		return fmt.Errorf("borrowing not found: %w", err)
	}

	// Check if already returned
	if borrowing.ReturnedAt != nil {
		return fmt.Errorf("game has already been returned")
	}

	// Update borrowing record with return date
	now := time.Now()
	borrowing.ReturnedAt = &now
	borrowing.IsOverdue = false // Clear overdue status on return

	if err := s.borrowingRepo.Update(borrowing); err != nil {
		return fmt.Errorf("failed to update borrowing record: %w", err)
	}

	// Update game availability
	game, err := s.gameRepo.GetByID(borrowing.GameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	game.IsAvailable = true
	if err := s.gameRepo.Update(game); err != nil {
		return fmt.Errorf("failed to update game availability: %w", err)
	}

	return nil
}

// GetOverdueItems retrieves all overdue borrowings
func (s *BorrowingService) GetOverdueItems() ([]*models.Borrowing, error) {
	borrowings, err := s.borrowingRepo.GetOverdue()
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue items: %w", err)
	}

	// Filter to only include items that are actually overdue and not returned
	var overdueItems []*models.Borrowing
	for _, borrowing := range borrowings {
		if borrowing.ReturnedAt == nil && borrowing.IsCurrentlyOverdue() {
			overdueItems = append(overdueItems, borrowing)
		}
	}

	return overdueItems, nil
}

// ExtendDueDate extends the due date for a borrowing
func (s *BorrowingService) ExtendDueDate(borrowingID int, newDueDate time.Time) error {
	if borrowingID <= 0 {
		return fmt.Errorf("invalid borrowing ID: %d", borrowingID)
	}

	// Get the borrowing record
	borrowing, err := s.borrowingRepo.GetByID(borrowingID)
	if err != nil {
		return fmt.Errorf("borrowing not found: %w", err)
	}

	// Check if already returned
	if borrowing.ReturnedAt != nil {
		return fmt.Errorf("cannot extend due date for returned item")
	}

	// Validate new due date
	if !newDueDate.After(borrowing.BorrowedAt) {
		return fmt.Errorf("new due date must be after borrowed date")
	}

	// Check if new due date is not too far in the future (max 90 days from borrowed date)
	maxDuration := 90 * 24 * time.Hour
	if newDueDate.Sub(borrowing.BorrowedAt) > maxDuration {
		return fmt.Errorf("due date cannot be more than 90 days from borrowed date")
	}

	// Update due date
	borrowing.DueDate = newDueDate
	borrowing.IsOverdue = borrowing.IsCurrentlyOverdue() // Recalculate overdue status

	if err := s.borrowingRepo.Update(borrowing); err != nil {
		return fmt.Errorf("failed to update borrowing: %w", err)
	}

	return nil
}

// GetBorrowingDetails retrieves detailed information about a borrowing
func (s *BorrowingService) GetBorrowingDetails(borrowingID int) (*models.Borrowing, error) {
	if borrowingID <= 0 {
		return nil, fmt.Errorf("invalid borrowing ID: %d", borrowingID)
	}

	borrowing, err := s.borrowingRepo.GetByID(borrowingID)
	if err != nil {
		return nil, fmt.Errorf("borrowing not found: %w", err)
	}

	return borrowing, nil
}

// GetActiveBorrowingsByUser retrieves all active borrowings for a user
func (s *BorrowingService) GetActiveBorrowingsByUser(userID int) ([]*models.Borrowing, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	activeBorrowings, err := s.borrowingRepo.GetActiveByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active borrowings: %w", err)
	}

	return activeBorrowings, nil
}

// GetBorrowingsByGame retrieves all borrowings for a specific game
func (s *BorrowingService) GetBorrowingsByGame(gameID int) ([]*models.Borrowing, error) {
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
		return nil, fmt.Errorf("failed to get game borrowings: %w", err)
	}

	return borrowings, nil
}

// UpdateOverdueStatus updates the overdue status for all active borrowings
func (s *BorrowingService) UpdateOverdueStatus() error {
	// Get all active borrowings
	allBorrowings, err := s.borrowingRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get all borrowings: %w", err)
	}

	// Update overdue status for active borrowings
	for _, borrowing := range allBorrowings {
		if borrowing.ReturnedAt == nil { // Only check active borrowings
			currentlyOverdue := borrowing.IsCurrentlyOverdue()
			if borrowing.IsOverdue != currentlyOverdue {
				borrowing.IsOverdue = currentlyOverdue
				if err := s.borrowingRepo.Update(borrowing); err != nil {
					return fmt.Errorf("failed to update borrowing %d overdue status: %w", borrowing.ID, err)
				}
			}
		}
	}

	return nil
}

// GetAllBorrowings retrieves all borrowings
func (s *BorrowingService) GetAllBorrowings() ([]*models.Borrowing, error) {
	borrowings, err := s.borrowingRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all borrowings: %w", err)
	}
	return borrowings, nil
}

// GetItemsDueSoon retrieves items that are due within the specified number of days
func (s *BorrowingService) GetItemsDueSoon(daysAhead int) ([]*models.Borrowing, error) {
	if daysAhead < 0 {
		return nil, fmt.Errorf("days ahead must be non-negative")
	}

	// Get all active borrowings
	allBorrowings, err := s.borrowingRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all borrowings: %w", err)
	}

	// Filter items due within the specified timeframe
	cutoffDate := time.Now().Add(time.Duration(daysAhead) * 24 * time.Hour)
	var itemsDueSoon []*models.Borrowing

	for _, borrowing := range allBorrowings {
		if borrowing.ReturnedAt == nil && // Not returned
			borrowing.DueDate.Before(cutoffDate) && // Due within timeframe
			!borrowing.IsCurrentlyOverdue() { // Not already overdue
			itemsDueSoon = append(itemsDueSoon, borrowing)
		}
	}

	return itemsDueSoon, nil
}