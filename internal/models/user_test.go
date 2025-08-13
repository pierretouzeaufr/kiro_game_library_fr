package models

import (
	"testing"
	"time"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: &User{
				ID:           1,
				Name:         "John Doe",
				Email:        "john.doe@example.com",
				RegisteredAt: time.Now(),
				IsActive:     true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			user: &User{
				Name:  "",
				Email: "john.doe@example.com",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too short",
			user: &User{
				Name:  "J",
				Email: "john.doe@example.com",
			},
			wantErr: true,
			errMsg:  "name must be at least 2 characters long",
		},
		{
			name: "name too long",
			user: &User{
				Name:  string(make([]byte, 101)), // 101 characters
				Email: "john.doe@example.com",
			},
			wantErr: true,
			errMsg:  "name must be less than 100 characters",
		},
		{
			name: "empty email",
			user: &User{
				Name:  "John Doe",
				Email: "",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format",
			user: &User{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "email too long",
			user: &User{
				Name:  "John Doe",
				Email: string(make([]byte, 250)) + "@example.com", // > 255 characters
			},
			wantErr: true,
			errMsg:  "email must be less than 255 characters",
		},
		{
			name: "valid email with plus sign",
			user: &User{
				Name:  "John Doe",
				Email: "john.doe+test@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid email with subdomain",
			user: &User{
				Name:  "John Doe",
				Email: "john.doe@mail.example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateUser() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateUserName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid name", "John Doe", false, ""},
		{"empty name", "", true, "name is required"},
		{"whitespace only", "   ", true, "name is required"},
		{"single character", "J", true, "name must be at least 2 characters long"},
		{"exactly 2 characters", "Jo", false, ""},
		{"exactly 100 characters", string(make([]byte, 100)), false, ""},
		{"over 100 characters", string(make([]byte, 101)), true, "name must be less than 100 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUserName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateUserName() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateUserEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid email", "test@example.com", false, ""},
		{"empty email", "", true, "email is required"},
		{"whitespace only", "   ", true, "email is required"},
		{"missing @", "testexample.com", true, "invalid email format"},
		{"missing domain", "test@", true, "invalid email format"},
		{"missing local part", "@example.com", true, "invalid email format"},
		{"no TLD", "test@example", true, "invalid email format"},
		{"valid with plus", "test+tag@example.com", false, ""},
		{"valid with subdomain", "test@mail.example.com", false, ""},
		{"valid with numbers", "test123@example123.com", false, ""},
		{"too long", string(make([]byte, 250)) + "@example.com", true, "email must be less than 255 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUserEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateUserEmail() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}