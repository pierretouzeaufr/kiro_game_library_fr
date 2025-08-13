package services

import (
	"board-game-library/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAll() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetBorrowingHistory(userID int) ([]*models.Borrowing, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

// MockBorrowingRepository is a mock implementation of BorrowingRepository
type MockBorrowingRepository struct {
	mock.Mock
}

func (m *MockBorrowingRepository) Create(borrowing *models.Borrowing) error {
	args := m.Called(borrowing)
	return args.Error(0)
}

func (m *MockBorrowingRepository) GetByID(id int) (*models.Borrowing, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingRepository) GetActiveByUser(userID int) ([]*models.Borrowing, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingRepository) GetByGame(gameID int) ([]*models.Borrowing, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingRepository) GetOverdue() ([]*models.Borrowing, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingRepository) Update(borrowing *models.Borrowing) error {
	args := m.Called(borrowing)
	return args.Error(0)
}

func (m *MockBorrowingRepository) ReturnGame(borrowingID int) error {
	args := m.Called(borrowingID)
	return args.Error(0)
}

func (m *MockBorrowingRepository) GetAll() ([]*models.Borrowing, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func TestNewUserService(t *testing.T) {
	userRepo := &MockUserRepository{}
	borrowingRepo := &MockBorrowingRepository{}

	service := NewUserService(userRepo, borrowingRepo)

	assert.NotNil(t, service)
	assert.Equal(t, userRepo, service.userRepo)
	assert.Equal(t, borrowingRepo, service.borrowingRepo)
}

func TestUserService_RegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		inputName     string
		inputEmail    string
		setupMocks    func(*MockUserRepository, *MockBorrowingRepository)
		expectedError string
	}{
		{
			name:       "successful registration",
			inputName:  "John Doe",
			inputEmail: "john@example.com",
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				userRepo.On("GetByEmail", "john@example.com").Return(nil, errors.New("not found"))
				userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:       "invalid name",
			inputName:  "",
			inputEmail: "john@example.com",
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "validation failed",
		},
		{
			name:       "invalid email",
			inputName:  "John Doe",
			inputEmail: "invalid-email",
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "validation failed",
		},
		{
			name:       "email already exists",
			inputName:  "John Doe",
			inputEmail: "john@example.com",
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				existingUser := &models.User{ID: 1, Email: "john@example.com"}
				userRepo.On("GetByEmail", "john@example.com").Return(existingUser, nil)
			},
			expectedError: "user with email john@example.com already exists",
		},
		{
			name:       "repository error",
			inputName:  "John Doe",
			inputEmail: "john@example.com",
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				userRepo.On("GetByEmail", "john@example.com").Return(nil, errors.New("not found"))
				userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(errors.New("database error"))
			},
			expectedError: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(userRepo, borrowingRepo)

			service := NewUserService(userRepo, borrowingRepo)
			user, err := service.RegisterUser(tt.inputName, tt.inputEmail)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.inputName, user.Name)
				assert.Equal(t, tt.inputEmail, user.Email)
				assert.True(t, user.IsActive)
			}

			userRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		setupMocks    func(*MockUserRepository, *MockBorrowingRepository)
		expectedError string
	}{
		{
			name:   "successful get user",
			userID: 1,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				user := &models.User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				userRepo.On("GetByID", 1).Return(user, nil)
			},
			expectedError: "",
		},
		{
			name:   "invalid user ID",
			userID: 0,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid user ID",
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				userRepo.On("GetByID", 999).Return(nil, errors.New("user not found"))
			},
			expectedError: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(userRepo, borrowingRepo)

			service := NewUserService(userRepo, borrowingRepo)
			user, err := service.GetUser(tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}

			userRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_CanUserBorrow(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		setupMocks    func(*MockUserRepository, *MockBorrowingRepository)
		expectedCan   bool
		expectedError string
	}{
		{
			name:   "user can borrow - no active borrowings",
			userID: 1,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: true}
				userRepo.On("GetByID", 1).Return(user, nil)
				borrowingRepo.On("GetActiveByUser", 1).Return([]*models.Borrowing{}, nil)
			},
			expectedCan:   true,
			expectedError: "",
		},
		{
			name:   "user can borrow - has active borrowings but not overdue",
			userID: 1,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: true}
				userRepo.On("GetByID", 1).Return(user, nil)
				
				// Create a borrowing that's not overdue
				futureDate := time.Now().Add(7 * 24 * time.Hour)
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-7 * 24 * time.Hour),
					DueDate:    futureDate,
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetActiveByUser", 1).Return([]*models.Borrowing{borrowing}, nil)
			},
			expectedCan:   true,
			expectedError: "",
		},
		{
			name:   "user cannot borrow - has overdue items",
			userID: 1,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: true}
				userRepo.On("GetByID", 1).Return(user, nil)
				
				// Create an overdue borrowing
				pastDate := time.Now().Add(-7 * 24 * time.Hour)
				borrowing := &models.Borrowing{
					ID:         1,
					UserID:     1,
					GameID:     1,
					BorrowedAt: time.Now().Add(-14 * 24 * time.Hour),
					DueDate:    pastDate,
					ReturnedAt: nil,
				}
				borrowingRepo.On("GetActiveByUser", 1).Return([]*models.Borrowing{borrowing}, nil)
			},
			expectedCan:   false,
			expectedError: "user has overdue items",
		},
		{
			name:   "user cannot borrow - inactive account",
			userID: 1,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				user := &models.User{ID: 1, Name: "John Doe", IsActive: false}
				userRepo.On("GetByID", 1).Return(user, nil)
			},
			expectedCan:   false,
			expectedError: "user account is inactive",
		},
		{
			name:   "invalid user ID",
			userID: 0,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedCan:   false,
			expectedError: "invalid user ID",
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				userRepo.On("GetByID", 999).Return(nil, errors.New("user not found"))
			},
			expectedCan:   false,
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(userRepo, borrowingRepo)

			service := NewUserService(userRepo, borrowingRepo)
			canBorrow, err := service.CanUserBorrow(tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, tt.expectedCan, canBorrow)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCan, canBorrow)
			}

			userRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserBorrowings(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		setupMocks    func(*MockUserRepository, *MockBorrowingRepository)
		expectedError string
	}{
		{
			name:   "successful get borrowings",
			userID: 1,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				user := &models.User{ID: 1, Name: "John Doe"}
				userRepo.On("GetByID", 1).Return(user, nil)
				
				borrowings := []*models.Borrowing{
					{ID: 1, UserID: 1, GameID: 1},
					{ID: 2, UserID: 1, GameID: 2},
				}
				userRepo.On("GetBorrowingHistory", 1).Return(borrowings, nil)
			},
			expectedError: "",
		},
		{
			name:   "invalid user ID",
			userID: 0,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid user ID",
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func(userRepo *MockUserRepository, borrowingRepo *MockBorrowingRepository) {
				userRepo.On("GetByID", 999).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(userRepo, borrowingRepo)

			service := NewUserService(userRepo, borrowingRepo)
			borrowings, err := service.GetUserBorrowings(tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, borrowings)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, borrowings)
				assert.Len(t, borrowings, 2)
			}

			userRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}