package models

import (
	"testing"
	"time"
)

func TestValidateBorrowing(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(14 * 24 * time.Hour)
	returnDate := now.Add(7 * 24 * time.Hour)

	tests := []struct {
		name      string
		borrowing *Borrowing
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid borrowing",
			borrowing: &Borrowing{
				ID:         1,
				UserID:     1,
				GameID:     1,
				BorrowedAt: now,
				DueDate:    futureDate,
				ReturnedAt: nil,
				IsOverdue:  false,
			},
			wantErr: false,
		},
		{
			name: "valid borrowing with return date",
			borrowing: &Borrowing{
				ID:         1,
				UserID:     1,
				GameID:     1,
				BorrowedAt: now,
				DueDate:    futureDate,
				ReturnedAt: &returnDate,
				IsOverdue:  false,
			},
			wantErr: false,
		},
		{
			name: "invalid user ID - zero",
			borrowing: &Borrowing{
				UserID:     0,
				GameID:     1,
				BorrowedAt: now,
				DueDate:    futureDate,
			},
			wantErr: true,
			errMsg:  "user ID must be a positive integer",
		},
		{
			name: "invalid user ID - negative",
			borrowing: &Borrowing{
				UserID:     -1,
				GameID:     1,
				BorrowedAt: now,
				DueDate:    futureDate,
			},
			wantErr: true,
			errMsg:  "user ID must be a positive integer",
		},
		{
			name: "invalid game ID - zero",
			borrowing: &Borrowing{
				UserID:     1,
				GameID:     0,
				BorrowedAt: now,
				DueDate:    futureDate,
			},
			wantErr: true,
			errMsg:  "game ID must be a positive integer",
		},
		{
			name: "due date before borrowed date",
			borrowing: &Borrowing{
				UserID:     1,
				GameID:     1,
				BorrowedAt: now,
				DueDate:    now.Add(-1 * time.Hour),
			},
			wantErr: true,
			errMsg:  "due date must be after borrowed date",
		},
		{
			name: "due date too far in future",
			borrowing: &Borrowing{
				UserID:     1,
				GameID:     1,
				BorrowedAt: now,
				DueDate:    now.Add(91 * 24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "due date cannot be more than 90 days from borrowed date",
		},
		{
			name: "return date before borrowed date",
			borrowing: &Borrowing{
				UserID:     1,
				GameID:     1,
				BorrowedAt: now,
				DueDate:    futureDate,
				ReturnedAt: &[]time.Time{now.Add(-1 * time.Hour)}[0],
			},
			wantErr: true,
			errMsg:  "return date must be after borrowed date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBorrowing(tt.borrowing)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBorrowing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateBorrowing() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateBorrowingUserID(t *testing.T) {
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
			err := validateBorrowingUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBorrowingUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateBorrowingUserID() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateBorrowingGameID(t *testing.T) {
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
			err := validateBorrowingGameID(tt.gameID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBorrowingGameID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateBorrowingGameID() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestIsCurrentlyOverdue(t *testing.T) {
	now := time.Now()
	pastDate := now.Add(-1 * time.Hour)
	futureDate := now.Add(1 * time.Hour)
	returnDate := now.Add(-30 * time.Minute)

	tests := []struct {
		name      string
		borrowing *Borrowing
		want      bool
	}{
		{
			name: "not overdue - future due date",
			borrowing: &Borrowing{
				DueDate:    futureDate,
				ReturnedAt: nil,
			},
			want: false,
		},
		{
			name: "overdue - past due date",
			borrowing: &Borrowing{
				DueDate:    pastDate,
				ReturnedAt: nil,
			},
			want: true,
		},
		{
			name: "not overdue - already returned",
			borrowing: &Borrowing{
				DueDate:    pastDate,
				ReturnedAt: &returnDate,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.borrowing.IsCurrentlyOverdue(); got != tt.want {
				t.Errorf("Borrowing.IsCurrentlyOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDaysOverdue(t *testing.T) {
	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)
	threeDaysAgo := now.Add(-72 * time.Hour)
	futureDate := now.Add(24 * time.Hour)
	returnDate := now.Add(-30 * time.Minute)

	tests := []struct {
		name      string
		borrowing *Borrowing
		want      int
	}{
		{
			name: "not overdue",
			borrowing: &Borrowing{
				DueDate:    futureDate,
				ReturnedAt: nil,
			},
			want: 0,
		},
		{
			name: "one day overdue",
			borrowing: &Borrowing{
				DueDate:    oneDayAgo,
				ReturnedAt: nil,
			},
			want: 1,
		},
		{
			name: "three days overdue",
			borrowing: &Borrowing{
				DueDate:    threeDaysAgo,
				ReturnedAt: nil,
			},
			want: 3,
		},
		{
			name: "already returned",
			borrowing: &Borrowing{
				DueDate:    threeDaysAgo,
				ReturnedAt: &returnDate,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.borrowing.DaysOverdue(); got != tt.want {
				t.Errorf("Borrowing.DaysOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}