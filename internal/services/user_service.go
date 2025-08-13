package services

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"fmt"
	"time"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo      repositories.UserRepository
	borrowingRepo repositories.BorrowingRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repositories.UserRepository, borrowingRepo repositories.BorrowingRepository) *UserService {
	return &UserService{
		userRepo:      userRepo,
		borrowingRepo: borrowingRepo,
	}
}

// RegisterUser creates a new user account
func (s *UserService) RegisterUser(name, email string) (*models.User, error) {
	// Create user model
	user := &models.User{
		Name:         name,
		Email:        email,
		RegisteredAt: time.Now(),
		IsActive:     true,
	}

	// Validate user data
	if err := models.ValidateUser(user); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Create user in repository
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id int) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	return users, nil
}

// GetUserBorrowings retrieves borrowing history for a user
func (s *UserService) GetUserBorrowings(userID int) ([]*models.Borrowing, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	borrowings, err := s.userRepo.GetBorrowingHistory(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user borrowings: %w", err)
	}

	return borrowings, nil
}

// CanUserBorrow checks if a user is eligible to borrow games
func (s *UserService) CanUserBorrow(userID int) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("invalid user ID: %d", userID)
	}

	// Verify user exists and is active
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return false, fmt.Errorf("user account is inactive")
	}

	// Check for active borrowings
	activeBorrowings, err := s.borrowingRepo.GetActiveByUser(userID)
	if err != nil {
		return false, fmt.Errorf("failed to check active borrowings: %w", err)
	}

	// Check if any active borrowings are overdue
	for _, borrowing := range activeBorrowings {
		if borrowing.IsCurrentlyOverdue() {
			return false, fmt.Errorf("user has overdue items")
		}
	}

	return true, nil
}

// GetActiveUserBorrowings retrieves current active borrowings for a user
func (s *UserService) GetActiveUserBorrowings(userID int) ([]*models.Borrowing, error) {
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

// UpdateUser updates user information
func (s *UserService) UpdateUser(user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if user.ID <= 0 {
		return fmt.Errorf("invalid user ID: %d", user.ID)
	}

	// Validate user data
	if err := models.ValidateUser(user); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if user exists
	_, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update user in repository
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser removes a user from the system
func (s *UserService) DeleteUser(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID: %d", userID)
	}
	
	// Check if user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// Check if user has active borrowings
	activeBorrowings, err := s.GetActiveUserBorrowings(userID)
	if err != nil {
		return fmt.Errorf("failed to check user borrowings: %w", err)
	}
	
	if len(activeBorrowings) > 0 {
		return fmt.Errorf("cannot delete user: has %d active borrowing(s)", len(activeBorrowings))
	}
	
	// Delete user from repository
	if err := s.userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}