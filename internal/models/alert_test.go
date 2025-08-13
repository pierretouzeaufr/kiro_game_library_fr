package models

import (
	"testing"
	"time"
)

func TestValidateAlert(t *testing.T) {
	tests := []struct {
		name    string
		alert   *Alert
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid alert",
			alert: &Alert{
				ID:        1,
				UserID:    1,
				GameID:    1,
				Type:      "overdue",
				Message:   "Game is overdue",
				CreatedAt: time.Now(),
				IsRead:    false,
			},
			wantErr: false,
		},
		{
			name: "valid reminder alert",
			alert: &Alert{
				UserID:  1,
				GameID:  1,
				Type:    "reminder",
				Message: "Game due in 2 days",
			},
			wantErr: false,
		},
		{
			name: "invalid user ID - zero",
			alert: &Alert{
				UserID:  0,
				GameID:  1,
				Type:    "overdue",
				Message: "Game is overdue",
			},
			wantErr: true,
			errMsg:  "user ID must be a positive integer",
		},
		{
			name: "invalid game ID - negative",
			alert: &Alert{
				UserID:  1,
				GameID:  -1,
				Type:    "overdue",
				Message: "Game is overdue",
			},
			wantErr: true,
			errMsg:  "game ID must be a positive integer",
		},
		{
			name: "invalid alert type",
			alert: &Alert{
				UserID:  1,
				GameID:  1,
				Type:    "invalid",
				Message: "Game is overdue",
			},
			wantErr: true,
			errMsg:  "invalid alert type: must be one of [overdue reminder]",
		},
		{
			name: "empty message",
			alert: &Alert{
				UserID:  1,
				GameID:  1,
				Type:    "overdue",
				Message: "",
			},
			wantErr: true,
			errMsg:  "alert message is required",
		},
		{
			name: "message too short",
			alert: &Alert{
				UserID:  1,
				GameID:  1,
				Type:    "overdue",
				Message: "Hi",
			},
			wantErr: true,
			errMsg:  "alert message must be at least 5 characters long",
		},
		{
			name: "message too long",
			alert: &Alert{
				UserID:  1,
				GameID:  1,
				Type:    "overdue",
				Message: string(make([]byte, 501)), // 501 characters
			},
			wantErr: true,
			errMsg:  "alert message must be less than 500 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAlert(tt.alert)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAlert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateAlert() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateAlertUserID(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		wantErr bool
		errMsg  string
	}{
		{"valid user ID", 1, false, ""},
		{"valid user ID - large number", 999999, false, ""},
		{"invalid user ID - zero", 0, true, "user ID must be a positive integer"},
		{"invalid user ID - negative", -1, true, "user ID must be a positive integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlertUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAlertUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateAlertUserID() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateAlertGameID(t *testing.T) {
	tests := []struct {
		name    string
		gameID  int
		wantErr bool
		errMsg  string
	}{
		{"valid game ID", 1, false, ""},
		{"valid game ID - large number", 999999, false, ""},
		{"invalid game ID - zero", 0, true, "game ID must be a positive integer"},
		{"invalid game ID - negative", -1, true, "game ID must be a positive integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlertGameID(tt.gameID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAlertGameID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateAlertGameID() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateAlertType(t *testing.T) {
	tests := []struct {
		name      string
		alertType string
		wantErr   bool
		errMsg    string
	}{
		{"valid type - overdue", "overdue", false, ""},
		{"valid type - reminder", "reminder", false, ""},
		{"valid type - case insensitive", "OVERDUE", false, ""},
		{"empty type", "", true, "alert type is required"},
		{"whitespace only", "   ", true, "alert type is required"},
		{"invalid type", "invalid", true, "invalid alert type: must be one of [overdue reminder]"},
		{"partial match", "over", true, "invalid alert type: must be one of [overdue reminder]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlertType(tt.alertType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAlertType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateAlertType() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateAlertMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
		errMsg  string
	}{
		{"valid message", "Game is overdue", false, ""},
		{"empty message", "", true, "alert message is required"},
		{"whitespace only", "   ", true, "alert message is required"},
		{"message too short", "Hi", true, "alert message must be at least 5 characters long"},
		{"exactly 5 characters", "Hello", false, ""},
		{"exactly 500 characters", string(make([]byte, 500)), false, ""},
		{"message too long", string(make([]byte, 501)), true, "alert message must be less than 500 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlertMessage(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAlertMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateAlertMessage() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}