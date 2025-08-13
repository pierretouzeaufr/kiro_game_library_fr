package models

import (
	"testing"
	"time"
)

func TestValidateGame(t *testing.T) {
	tests := []struct {
		name    string
		game    *Game
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid game",
			game: &Game{
				ID:          1,
				Name:        "Monopoly",
				Description: "Classic board game",
				Category:    "Strategy",
				EntryDate:   time.Now(),
				Condition:   "good",
				IsAvailable: true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			game: &Game{
				Name:        "",
				Description: "Classic board game",
				Category:    "Strategy",
				Condition:   "good",
			},
			wantErr: true,
			errMsg:  "game name is required",
		},
		{
			name: "invalid condition",
			game: &Game{
				Name:        "Monopoly",
				Description: "Classic board game",
				Category:    "Strategy",
				Condition:   "terrible",
			},
			wantErr: true,
			errMsg:  "invalid game condition: must be one of [excellent good fair poor]",
		},
		{
			name: "description too long",
			game: &Game{
				Name:        "Monopoly",
				Description: string(make([]byte, 1001)), // 1001 characters
				Category:    "Strategy",
				Condition:   "good",
			},
			wantErr: true,
			errMsg:  "game description must be less than 1000 characters",
		},
		{
			name: "valid with empty description",
			game: &Game{
				Name:      "Monopoly",
				Category:  "Strategy",
				Condition: "excellent",
			},
			wantErr: false,
		},
		{
			name: "valid with empty category",
			game: &Game{
				Name:        "Monopoly",
				Description: "Classic board game",
				Condition:   "fair",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGame(tt.game)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateGame() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateGameName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid name", "Monopoly", false, ""},
		{"empty name", "", true, "game name is required"},
		{"whitespace only", "   ", true, "game name is required"},
		{"single character", "M", true, "game name must be at least 2 characters long"},
		{"exactly 2 characters", "Go", false, ""},
		{"exactly 200 characters", string(make([]byte, 200)), false, ""},
		{"over 200 characters", string(make([]byte, 201)), true, "game name must be less than 200 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGameName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGameName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateGameName() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateGameDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid description", "A classic board game", false, ""},
		{"empty description", "", false, ""},
		{"whitespace only", "   ", false, ""},
		{"exactly 1000 characters", string(make([]byte, 1000)), false, ""},
		{"over 1000 characters", string(make([]byte, 1001)), true, "game description must be less than 1000 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGameDescription(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGameDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateGameDescription() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateGameCategory(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid category", "Strategy", false, ""},
		{"empty category", "", false, ""},
		{"whitespace only", "   ", false, ""},
		{"exactly 100 characters", string(make([]byte, 100)), false, ""},
		{"over 100 characters", string(make([]byte, 101)), true, "game category must be less than 100 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGameCategory(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGameCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateGameCategory() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateGameCondition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid condition - excellent", "excellent", false, ""},
		{"valid condition - good", "good", false, ""},
		{"valid condition - fair", "fair", false, ""},
		{"valid condition - poor", "poor", false, ""},
		{"valid condition - case insensitive", "EXCELLENT", false, ""},
		{"empty condition", "", true, "game condition is required"},
		{"whitespace only", "   ", true, "game condition is required"},
		{"invalid condition", "terrible", true, "invalid game condition: must be one of [excellent good fair poor]"},
		{"invalid condition - partial match", "goo", true, "invalid game condition: must be one of [excellent good fair poor]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGameCondition(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGameCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateGameCondition() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}