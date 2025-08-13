package models

import (
	"fmt"
	"time"
)

// Borrowing represents a game borrowing record
type Borrowing struct {
	ID         int        `json:"id" db:"id"`
	UserID     int        `json:"user_id" db:"user_id"`
	GameID     int        `json:"game_id" db:"game_id"`
	BorrowedAt time.Time  `json:"borrowed_at" db:"borrowed_at"`
	DueDate    time.Time  `json:"due_date" db:"due_date"`
	ReturnedAt *time.Time `json:"returned_at" db:"returned_at"`
	IsOverdue  bool       `json:"is_overdue" db:"is_overdue"`
}

// ValidateBorrowing validates a Borrowing struct
func ValidateBorrowing(borrowing *Borrowing) error {
	if err := validateBorrowingUserID(borrowing.UserID); err != nil {
		return err
	}
	
	if err := validateBorrowingGameID(borrowing.GameID); err != nil {
		return err
	}
	
	if err := validateBorrowingDates(borrowing.BorrowedAt, borrowing.DueDate, borrowing.ReturnedAt); err != nil {
		return err
	}
	
	return nil
}

// validateBorrowingUserID validates the user ID field
func validateBorrowingUserID(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("user ID must be a positive integer")
	}
	
	return nil
}

// validateBorrowingGameID validates the game ID field
func validateBorrowingGameID(gameID int) error {
	if gameID <= 0 {
		return fmt.Errorf("game ID must be a positive integer")
	}
	
	return nil
}

// validateBorrowingDates validates the borrowing date logic
func validateBorrowingDates(borrowedAt, dueDate time.Time, returnedAt *time.Time) error {
	// Check if due date is after borrowed date
	if !dueDate.After(borrowedAt) {
		return fmt.Errorf("due date must be after borrowed date")
	}
	
	// Check if due date is not too far in the future (max 90 days)
	maxDuration := 90 * 24 * time.Hour
	if dueDate.Sub(borrowedAt) > maxDuration {
		return fmt.Errorf("due date cannot be more than 90 days from borrowed date")
	}
	
	// If returned, check that return date is after borrowed date
	if returnedAt != nil {
		if !returnedAt.After(borrowedAt) {
			return fmt.Errorf("return date must be after borrowed date")
		}
	}
	
	return nil
}

// IsCurrentlyOverdue checks if the borrowing is currently overdue
func (b *Borrowing) IsCurrentlyOverdue() bool {
	// If already returned, not overdue
	if b.ReturnedAt != nil {
		return false
	}
	
	// Check if current time is past due date
	return time.Now().After(b.DueDate)
}

// DaysOverdue returns the number of days the item is overdue (0 if not overdue)
func (b *Borrowing) DaysOverdue() int {
	if !b.IsCurrentlyOverdue() {
		return 0
	}
	
	duration := time.Since(b.DueDate)
	return int(duration.Hours() / 24)
}