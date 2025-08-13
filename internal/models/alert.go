package models

import (
	"fmt"
	"strings"
	"time"
)

// Alert represents a system alert for overdue items or reminders
type Alert struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	GameID    int       `json:"game_id" db:"game_id"`
	Type      string    `json:"type" db:"type"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsRead    bool      `json:"is_read" db:"is_read"`
}

// ValidAlertTypes defines the allowed alert types
var ValidAlertTypes = []string{"overdue", "reminder", "custom"}

// ValidateAlert validates an Alert struct
func ValidateAlert(alert *Alert) error {
	if err := validateAlertUserID(alert.UserID); err != nil {
		return err
	}
	
	if err := validateAlertGameID(alert.GameID); err != nil {
		return err
	}
	
	if err := validateAlertType(alert.Type); err != nil {
		return err
	}
	
	if err := validateAlertMessage(alert.Message); err != nil {
		return err
	}
	
	return nil
}

// validateAlertUserID validates the user ID field
func validateAlertUserID(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("user ID must be a positive integer")
	}
	
	return nil
}

// validateAlertGameID validates the game ID field
func validateAlertGameID(gameID int) error {
	if gameID <= 0 {
		return fmt.Errorf("game ID must be a positive integer")
	}
	
	return nil
}

// validateAlertType validates the alert type field
func validateAlertType(alertType string) error {
	alertType = strings.TrimSpace(alertType)
	if alertType == "" {
		return fmt.Errorf("alert type is required")
	}
	
	// Check if type is in valid types list
	for _, validType := range ValidAlertTypes {
		if strings.EqualFold(alertType, validType) {
			return nil
		}
	}
	
	return fmt.Errorf("invalid alert type: must be one of %v", ValidAlertTypes)
}

// validateAlertMessage validates the alert message field
func validateAlertMessage(message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("alert message is required")
	}
	
	if len(message) < 5 {
		return fmt.Errorf("alert message must be at least 5 characters long")
	}
	
	if len(message) > 500 {
		return fmt.Errorf("alert message must be less than 500 characters")
	}
	
	return nil
}