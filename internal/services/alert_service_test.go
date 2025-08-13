package services

import (
	"board-game-library/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAlertRepository is a mock implementation of AlertRepository
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(alert *models.Alert) error {
	args := m.Called(alert)
	return args.Error(0)
}

func (m *MockAlertRepository) GetByID(id int) (*models.Alert, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetUnread() ([]*models.Alert, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetByUser(userID int) ([]*models.Alert, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAll() ([]*models.Alert, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) MarkAsRead(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAlertRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestNewAlertService(t *testing.T) {
	alertRepo := &MockAlertRepository{}
	borrowingRepo := &MockBorrowingRepository{}
	userRepo := &MockUserRepository{}
	gameRepo := &MockGameRepository{}

	service := NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	assert.NotNil(t, service)
	assert.Equal(t, alertRepo, service.alertRepo)
	assert.Equal(t, borrowingRepo, service.borrowingRepo)
	assert.Equal(t, userRepo, service.userRepo)
	assert.Equal(t, gameRepo, service.gameRepo)
}

func TestAlertService_GenerateOverdueAlerts(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockAlertRepository, *MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedError string
	}{
		{
			name: "successful overdue alert generation",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// Create an overdue borrowing
				pastDate := time.Now().Add(-7 * 24 * time.Hour)
				overdueBorrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-14 * 24 * time.Hour),
					DueDate:    pastDate,
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetOverdue").Return([]*models.Borrowing{overdueBorrowing}, nil)
				
				// No existing alerts
				alertRepo.On("GetByUser", 1).Return([]*models.Alert{}, nil)
				
				// Game details
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
				
				// Create alert
				alertRepo.On("Create", mock.AnythingOfType("*models.Alert")).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "no overdue items",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowingRepo.On("GetOverdue").Return([]*models.Borrowing{}, nil)
			},
			expectedError: "",
		},
		{
			name: "existing alert prevents duplicate",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// Create an overdue borrowing
				pastDate := time.Now().Add(-7 * 24 * time.Hour)
				overdueBorrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-14 * 24 * time.Hour),
					DueDate:    pastDate,
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetOverdue").Return([]*models.Borrowing{overdueBorrowing}, nil)
				
				// Existing overdue alert
				existingAlert := &models.Alert{
					ID:     1,
					UserID: 1,
					GameID: 1,
					Type:   "overdue",
					IsRead: false,
				}
				alertRepo.On("GetByUser", 1).Return([]*models.Alert{existingAlert}, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowingRepo.On("GetOverdue").Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get overdue borrowings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertRepo := &MockAlertRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(alertRepo, borrowingRepo, userRepo, gameRepo)

			service := NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)
			err := service.GenerateOverdueAlerts()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			alertRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestAlertService_GenerateReminderAlerts(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockAlertRepository, *MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedError string
	}{
		{
			name: "successful reminder alert generation",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// Create a borrowing due in 1 day
				futureDate := time.Now().Add(1 * 24 * time.Hour)
				dueSoonBorrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-13 * 24 * time.Hour),
					DueDate:    futureDate,
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetAll").Return([]*models.Borrowing{dueSoonBorrowing}, nil)
				
				// No existing alerts
				alertRepo.On("GetByUser", 1).Return([]*models.Alert{}, nil)
				
				// Game details
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
				
				// Create alert
				alertRepo.On("Create", mock.AnythingOfType("*models.Alert")).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "no items due soon",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// Create a borrowing due in 5 days (outside 2-day window)
				futureDate := time.Now().Add(5 * 24 * time.Hour)
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-9 * 24 * time.Hour),
					DueDate:    futureDate,
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetAll").Return([]*models.Borrowing{borrowing}, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				borrowingRepo.On("GetAll").Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get all borrowings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertRepo := &MockAlertRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(alertRepo, borrowingRepo, userRepo, gameRepo)

			service := NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)
			err := service.GenerateReminderAlerts()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			alertRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestAlertService_GetActiveAlerts(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockAlertRepository, *MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedCount int
		expectedError string
	}{
		{
			name: "successful get active alerts",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				alerts := []*models.Alert{
					{ID: 1, UserID: 1, GameID: 1, Type: "overdue", IsRead: false},
					{ID: 2, UserID: 2, GameID: 2, Type: "reminder", IsRead: false},
				}
				alertRepo.On("GetUnread").Return(alerts, nil)
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name: "repository error",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				alertRepo.On("GetUnread").Return(nil, errors.New("database error"))
			},
			expectedCount: 0,
			expectedError: "failed to get active alerts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertRepo := &MockAlertRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(alertRepo, borrowingRepo, userRepo, gameRepo)

			service := NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)
			alerts, err := service.GetActiveAlerts()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, alerts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, alerts)
				assert.Len(t, alerts, tt.expectedCount)
			}

			alertRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestAlertService_MarkAlertAsRead(t *testing.T) {
	tests := []struct {
		name          string
		alertID       int
		setupMocks    func(*MockAlertRepository, *MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedError string
	}{
		{
			name:    "successful mark as read",
			alertID: 1,
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				alert := &models.Alert{ID: 1, UserID: 1, GameID: 1, Type: "overdue"}
				alertRepo.On("GetByID", 1).Return(alert, nil)
				alertRepo.On("MarkAsRead", 1).Return(nil)
			},
			expectedError: "",
		},
		{
			name:    "invalid alert ID",
			alertID: 0,
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid alert ID",
		},
		{
			name:    "alert not found",
			alertID: 999,
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				alertRepo.On("GetByID", 999).Return(nil, errors.New("alert not found"))
			},
			expectedError: "alert not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertRepo := &MockAlertRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(alertRepo, borrowingRepo, userRepo, gameRepo)

			service := NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)
			err := service.MarkAlertAsRead(tt.alertID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			alertRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestAlertService_GetAlertsSummaryByUser(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockAlertRepository, *MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedUsers int
		expectedError string
	}{
		{
			name: "successful summary generation",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				alerts := []*models.Alert{
					{ID: 1, UserID: 1, GameID: 1, Type: "overdue", IsRead: false},
					{ID: 2, UserID: 1, GameID: 2, Type: "reminder", IsRead: false},
					{ID: 3, UserID: 2, GameID: 3, Type: "overdue", IsRead: false},
				}
				alertRepo.On("GetUnread").Return(alerts, nil)
			},
			expectedUsers: 2,
			expectedError: "",
		},
		{
			name: "no alerts",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				alertRepo.On("GetUnread").Return([]*models.Alert{}, nil)
			},
			expectedUsers: 0,
			expectedError: "",
		},
		{
			name: "repository error",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				alertRepo.On("GetUnread").Return(nil, errors.New("database error"))
			},
			expectedUsers: 0,
			expectedError: "failed to get unread alerts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertRepo := &MockAlertRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(alertRepo, borrowingRepo, userRepo, gameRepo)

			service := NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)
			summary, err := service.GetAlertsSummaryByUser()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, summary)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, summary)
				assert.Len(t, summary, tt.expectedUsers)
				
				// Verify summary structure for first test case
				if tt.expectedUsers == 2 {
					user1Summary := summary[1]
					assert.Equal(t, 1, user1Summary.UserID)
					assert.Equal(t, 2, user1Summary.TotalAlerts)
					assert.Equal(t, 1, user1Summary.OverdueCount)
					assert.Equal(t, 1, user1Summary.ReminderCount)
					
					user2Summary := summary[2]
					assert.Equal(t, 2, user2Summary.UserID)
					assert.Equal(t, 1, user2Summary.TotalAlerts)
					assert.Equal(t, 1, user2Summary.OverdueCount)
					assert.Equal(t, 0, user2Summary.ReminderCount)
				}
			}

			alertRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}

func TestAlertService_CreateCustomAlert(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		gameID        int
		alertType     string
		message       string
		setupMocks    func(*MockAlertRepository, *MockBorrowingRepository, *MockUserRepository, *MockGameRepository)
		expectedError string
	}{
		{
			name:      "successful custom alert creation",
			userID:    1,
			gameID:    1,
			alertType: "reminder",
			message:   "Custom reminder message for testing",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				user := &models.User{ID: 1, Name: "John Doe"}
				userRepo.On("GetByID", 1).Return(user, nil)
				
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
				
				alertRepo.On("Create", mock.AnythingOfType("*models.Alert")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:      "invalid user ID",
			userID:    0,
			gameID:    1,
			alertType: "reminder",
			message:   "Test message",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid user ID",
		},
		{
			name:      "invalid game ID",
			userID:    1,
			gameID:    0,
			alertType: "reminder",
			message:   "Test message",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid game ID",
		},
		{
			name:      "user not found",
			userID:    999,
			gameID:    1,
			alertType: "reminder",
			message:   "Test message",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				userRepo.On("GetByID", 999).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
		{
			name:      "invalid alert type",
			userID:    1,
			gameID:    1,
			alertType: "invalid",
			message:   "Test message",
			setupMocks: func(alertRepo *MockAlertRepository, borrowingRepo *MockBorrowingRepository, userRepo *MockUserRepository, gameRepo *MockGameRepository) {
				user := &models.User{ID: 1, Name: "John Doe"}
				userRepo.On("GetByID", 1).Return(user, nil)
				
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
			},
			expectedError: "alert validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertRepo := &MockAlertRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			userRepo := &MockUserRepository{}
			gameRepo := &MockGameRepository{}
			tt.setupMocks(alertRepo, borrowingRepo, userRepo, gameRepo)

			service := NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)
			alert, err := service.CreateCustomAlert(tt.userID, tt.gameID, tt.alertType, tt.message)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, alert)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, alert)
				assert.Equal(t, tt.userID, alert.UserID)
				assert.Equal(t, tt.gameID, alert.GameID)
				assert.Equal(t, tt.alertType, alert.Type)
				assert.Equal(t, tt.message, alert.Message)
				assert.False(t, alert.IsRead)
			}

			alertRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			gameRepo.AssertExpectations(t)
		})
	}
}