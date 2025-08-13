package services

import (
	"board-game-library/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewBorrowingService(t *testing.T) {
	borrowingRepo := &MockBorrowingRepository{}
	userRepo := &MockUserRepository{}
	gameRepo := &MockGameRepository{}

	service := NewBorrowingService(borrowingRepo, userRepo, gameRepo)

	assert.NotNil(t, service)
	assert.Equal(t, borrowingRepo, service.borrowingRepo)
	assert.Equal(t, userRepo, service.userRepo)
	assert.Equal(t, gameRepo, service.gameRepo)
}

func TestBorrowingService_BorrowGame(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		gameID        int
		dueDate       time.Time
		setupMocks    func(*MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedError string
	}{
		{
			name:    "successful borrowing",
			userID:  1,
			gameID:  1,
			dueDate: time.Now().Add(14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: true}
				userRepo.On("GetByID", 1).Return(user, nil)
				borrowingRepo.On("GetActiveByUser", 1).Return([]*models.Borrowing{}, nil)
				
				game := &models.Game{ID: 1, Name: "Monopoly", IsAvailable: true}
				gameRepo.On("GetByID", 1).Return(game, nil)
				borrowingRepo.On("Create", mock.AnythingOfType("*models.Borrowing")).Return(nil)
				gameRepo.On("Update", mock.MatchedBy(func(g *models.Game) bool {
					return g.ID == 1 && !g.IsAvailable
				})).Return(nil)
			},
			expectedError: "",
		},
		{
			name:    "invalid user ID",
			userID:  0,
			gameID:  1,
			dueDate: time.Now().Add(14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid user ID",
		},
		{
			name:    "invalid game ID",
			userID:  1,
			gameID:  0,
			dueDate: time.Now().Add(14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid game ID",
		},
		{
			name:    "user not found",
			userID:  999,
			gameID:  1,
			dueDate: time.Now().Add(14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				userRepo.On("GetByID", 999).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
		{
			name:    "inactive user",
			userID:  1,
			gameID:  1,
			dueDate: time.Now().Add(14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: false}
				userRepo.On("GetByID", 1).Return(user, nil)
			},
			expectedError: "user account is inactive",
		},
		{
			name:    "user has overdue items",
			userID:  1,
			gameID:  1,
			dueDate: time.Now().Add(14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: true}
				userRepo.On("GetByID", 1).Return(user, nil)
				
				// Create an overdue borrowing
				pastDate := time.Now().Add(-7 * 24 * time.Hour)
				overdueBorrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     2,
					BorrowedAt: time.Now().Add(-14 * 24 * time.Hour),
					DueDate:    pastDate,
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetActiveByUser", 1).Return([]*models.Borrowing{overdueBorrowing}, nil)
			},
			expectedError: "user has overdue items and cannot borrow",
		},
		{
			name:    "game not available",
			userID:  1,
			gameID:  1,
			dueDate: time.Now().Add(14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: true}
				userRepo.On("GetByID", 1).Return(user, nil)
				borrowingRepo.On("GetActiveByUser", 1).Return([]*models.Borrowing{}, nil)
				
				game := &models.Game{ID: 1, Name: "Monopoly", IsAvailable: false}
				gameRepo.On("GetByID", 1).Return(game, nil)
			},
			expectedError: "game is not available for borrowing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(borrowingRepo, userRepo, gameRepo)

			service := NewBorrowingService(borrowingRepo, userRepo, gameRepo)
			borrowing, err := service.BorrowGame(tt.userID, tt.gameID, tt.dueDate)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, borrowing)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, borrowing)
				assert.Equal(t, tt.userID, borrowing.UserID)
				assert.Equal(t, tt.gameID, borrowing.GameID)
				assert.Nil(t, borrowing.ReturnedAt)
			}

			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestBorrowingService_BorrowGameWithDefaultDueDate(t *testing.T) {
	borrowingRepo := &MockBorrowingRepository{}
	userRepo := &MockUserRepository{}
	gameRepo := &MockGameRepository{}

	user := &models.User{ID: 1, Name: "John Doe", IsActive: true}
	userRepo.On("GetByID", 1).Return(user, nil)
	borrowingRepo.On("GetActiveByUser", 1).Return([]*models.Borrowing{}, nil)
	
	game := &models.Game{ID: 1, Name: "Monopoly", IsAvailable: true}
	gameRepo.On("GetByID", 1).Return(game, nil)
	borrowingRepo.On("Create", mock.AnythingOfType("*models.Borrowing")).Return(nil)
	gameRepo.On("Update", mock.AnythingOfType("*models.Game")).Return(nil)

	service := NewBorrowingService(borrowingRepo, userRepo, gameRepo)
	borrowing, err := service.BorrowGameWithDefaultDueDate(1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, borrowing)
	
	// Check that due date is approximately 14 days from now
	expectedDueDate := time.Now().Add(14 * 24 * time.Hour)
	assert.WithinDuration(t, expectedDueDate, borrowing.DueDate, time.Minute)

	borrowingRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	gameRepo.AssertExpectations(t)
}

func TestBorrowingService_ReturnGame(t *testing.T) {
	tests := []struct {
		name          string
		borrowingID   int
		setupMocks    func(*MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedError string
	}{
		{
			name:        "successful return",
			borrowingID: 1,
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
					DueDate:    time.Now().Add(7 * 24 * time.Hour),
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetByID", 1).Return(borrowing, nil)
				borrowingRepo.On("Update", mock.MatchedBy(func(b *models.Borrowing) bool {
					return b.ID == 1 && b.ReturnedAt != nil
				})).Return(nil)
				
				game := &models.Game{ID: 1, Name: "Monopoly", IsAvailable: false}
				gameRepo.On("GetByID", 1).Return(game, nil)
				gameRepo.On("Update", mock.MatchedBy(func(g *models.Game) bool {
					return g.ID == 1 && g.IsAvailable
				})).Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "invalid borrowing ID",
			borrowingID: 0,
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid borrowing ID",
		},
		{
			name:        "borrowing not found",
			borrowingID: 999,
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowingRepo.On("GetByID", 999).Return(nil, errors.New("borrowing not found"))
			},
			expectedError: "borrowing not found",
		},
		{
			name:        "already returned",
			borrowingID: 1,
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				returnTime := time.Now().Add(-1 * time.Hour)
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
					DueDate:    time.Now().Add(7 * 24 * time.Hour),
					ReturnedAt: &returnTime,
				}
				borrowingRepo.On("GetByID", 1).Return(borrowing, nil)
			},
			expectedError: "game has already been returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(borrowingRepo, userRepo, gameRepo)

			service := NewBorrowingService(borrowingRepo, userRepo, gameRepo)
			err := service.ReturnGame(tt.borrowingID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestBorrowingService_GetOverdueItems(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedCount int
		expectedError string
	}{
		{
			name: "successful get overdue items",
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				pastDate := time.Now().Add(-7 * 24 * time.Hour)
				futureDate := time.Now().Add(7 * 24 * time.Hour)
				returnTime := time.Now().Add(-1 * time.Hour)
				
				borrowings := []*models.Borrowing{
					{ID: 1, UserID: 1, GameID: 1, DueDate: pastDate, ReturnedAt: nil}, // Overdue and active
					{ID: 2, UserID: 2, GameID: 2, DueDate: futureDate, ReturnedAt: nil}, // Not overdue
					{ID: 3, UserID: 3, GameID: 3, DueDate: pastDate, ReturnedAt: &returnTime}, // Overdue but returned
				}
				borrowingRepo.On("GetOverdue").Return(borrowings, nil)
			},
			expectedCount: 1,
			expectedError: "",
		},
		{
			name: "repository error",
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowingRepo.On("GetOverdue").Return(nil, errors.New("database error"))
			},
			expectedCount: 0,
			expectedError: "failed to get overdue items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(borrowingRepo, userRepo, gameRepo)

			service := NewBorrowingService(borrowingRepo, userRepo, gameRepo)
			items, err := service.GetOverdueItems()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, items)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, items)
				assert.Len(t, items, tt.expectedCount)
			}

			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestBorrowingService_ExtendDueDate(t *testing.T) {
	tests := []struct {
		name          string
		borrowingID   int
		newDueDate    time.Time
		setupMocks    func(*MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedError string
	}{
		{
			name:        "successful extension",
			borrowingID: 1,
			newDueDate:  time.Now().Add(21 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
					DueDate:    time.Now().Add(7 * 24 * time.Hour),
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetByID", 1).Return(borrowing, nil)
				borrowingRepo.On("Update", mock.AnythingOfType("*models.Borrowing")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "invalid borrowing ID",
			borrowingID: 0,
			newDueDate:  time.Now().Add(21 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid borrowing ID",
		},
		{
			name:        "already returned",
			borrowingID: 1,
			newDueDate:  time.Now().Add(21 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				returnTime := time.Now().Add(-1 * time.Hour)
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
					DueDate:    time.Now().Add(7 * 24 * time.Hour),
					ReturnedAt: &returnTime,
				}
				borrowingRepo.On("GetByID", 1).Return(borrowing, nil)
			},
			expectedError: "cannot extend due date for returned item",
		},
		{
			name:        "new due date before borrowed date",
			borrowingID: 1,
			newDueDate:  time.Now().Add(-14 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
					DueDate:    time.Now().Add(7 * 24 * time.Hour),
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetByID", 1).Return(borrowing, nil)
			},
			expectedError: "new due date must be after borrowed date",
		},
		{
			name:        "new due date too far in future",
			borrowingID: 1,
			newDueDate:  time.Now().Add(100 * 24 * time.Hour),
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
					DueDate:    time.Now().Add(7 * 24 * time.Hour),
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetByID", 1).Return(borrowing, nil)
			},
			expectedError: "due date cannot be more than 90 days from borrowed date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(borrowingRepo, userRepo, gameRepo)

			service := NewBorrowingService(borrowingRepo, userRepo, gameRepo)
			err := service.ExtendDueDate(tt.borrowingID, tt.newDueDate)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestBorrowingService_GetItemsDueSoon(t *testing.T) {
	tests := []struct {
		name          string
		daysAhead     int
		setupMocks    func(*MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedCount int
		expectedError string
	}{
		{
			name:      "successful get items due soon",
			daysAhead: 2,
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				now := time.Now()
				borrowings := []*models.Borrowing{
					{ID: 1, UserID: 1, GameID: 1, DueDate: now.Add(1 * 24 * time.Hour), ReturnedAt: nil}, // Due in 1 day
					{ID: 2, UserID: 2, GameID: 2, DueDate: now.Add(3 * 24 * time.Hour), ReturnedAt: nil}, // Due in 3 days (outside range)
					{ID: 3, UserID: 3, GameID: 3, DueDate: now.Add(-1 * 24 * time.Hour), ReturnedAt: nil}, // Overdue
					{ID: 4, UserID: 4, GameID: 4, DueDate: now.Add(1 * 24 * time.Hour), ReturnedAt: &now}, // Due soon but returned
				}
				borrowingRepo.On("GetAll").Return(borrowings, nil)
			},
			expectedCount: 1,
			expectedError: "",
		},
		{
			name:      "negative days ahead",
			daysAhead: -1,
			setupMocks: func(borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedCount: 0,
			expectedError: "days ahead must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(borrowingRepo, userRepo, gameRepo)

			service := NewBorrowingService(borrowingRepo, userRepo, gameRepo)
			items, err := service.GetItemsDueSoon(tt.daysAhead)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, items)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, items)
				assert.Len(t, items, tt.expectedCount)
			}

			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}