package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// User represents a library user
type User struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CurrentLoans int       `json:"current_loans" db:"-"` // Not stored in DB, calculated at runtime
}

// ValidateUser validates a User struct
func ValidateUser(user *User) error {
	if err := validateUserName(user.Name); err != nil {
		return err
	}
	
	if err := validateUserEmail(user.Email); err != nil {
		return err
	}
	
	return nil
}

// validateUserName validates the user name field
func validateUserName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}
	
	if len(name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}
	
	return nil
}

// validateUserEmail validates the user email field
func validateUserEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	
	if len(email) > 255 {
		return fmt.Errorf("email must be less than 255 characters")
	}
	
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}