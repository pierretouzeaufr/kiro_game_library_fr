package models

import (
	"fmt"
	"strings"
	"time"
)

// Game represents a board game in the library
type Game struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	EntryDate   time.Time `json:"entry_date" db:"entry_date"`
	Condition   string    `json:"condition" db:"condition"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
}

// ValidConditions defines the allowed condition values
var ValidConditions = []string{"excellent", "good", "fair", "poor"}

// ValidateGame validates a Game struct
func ValidateGame(game *Game) error {
	if err := validateGameName(game.Name); err != nil {
		return err
	}
	
	if err := validateGameDescription(game.Description); err != nil {
		return err
	}
	
	if err := validateGameCategory(game.Category); err != nil {
		return err
	}
	
	if err := validateGameCondition(game.Condition); err != nil {
		return err
	}
	
	return nil
}

// validateGameName validates the game name field
func validateGameName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("game name is required")
	}
	
	if len(name) < 2 {
		return fmt.Errorf("game name must be at least 2 characters long")
	}
	
	if len(name) > 200 {
		return fmt.Errorf("game name must be less than 200 characters")
	}
	
	return nil
}

// validateGameDescription validates the game description field
func validateGameDescription(description string) error {
	description = strings.TrimSpace(description)
	
	if len(description) > 1000 {
		return fmt.Errorf("game description must be less than 1000 characters")
	}
	
	return nil
}

// validateGameCategory validates the game category field
func validateGameCategory(category string) error {
	category = strings.TrimSpace(category)
	
	if len(category) > 100 {
		return fmt.Errorf("game category must be less than 100 characters")
	}
	
	return nil
}

// validateGameCondition validates the game condition field
func validateGameCondition(condition string) error {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return fmt.Errorf("game condition is required")
	}
	
	// Check if condition is in valid conditions list
	for _, validCondition := range ValidConditions {
		if strings.EqualFold(condition, validCondition) {
			return nil
		}
	}
	
	return fmt.Errorf("invalid game condition: must be one of %v", ValidConditions)
}