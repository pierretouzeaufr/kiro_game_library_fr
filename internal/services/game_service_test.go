package services

import (
	"board-game-library/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGameRepository is a mock implementation of GameRepository
type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) Create(game *models.Game) error {
	args := m.Called(game)
	return args.Error(0)
}

func (m *MockGameRepository) GetByID(id int) (*models.Game, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Game), args.Error(1)
}

func (m *MockGameRepository) GetAll() ([]*models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameRepository) Search(query string) ([]*models.Game, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameRepository) Update(game *models.Game) error {
	args := m.Called(game)
	return args.Error(0)
}

func (m *MockGameRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGameRepository) GetAvailable() ([]*models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func TestNewGameService(t *testing.T) {
	gameRepo := &MockGameRepository{}
	borrowingRepo := &MockBorrowingRepository{}

	service := NewGameService(gameRepo, borrowingRepo)

	assert.NotNil(t, service)
	assert.Equal(t, gameRepo, service.gameRepo)
	assert.Equal(t, borrowingRepo, service.borrowingRepo)
}

func TestGameService_AddGame(t *testing.T) {
	tests := []struct {
		name          string
		inputName     string
		inputDesc     string
		inputCategory string
		inputCondition string
		setupMocks    func(*MockGameRepository, *MockBorrowingRepository)
		expectedError string
	}{
		{
			name:           "successful game creation",
			inputName:      "Monopoly",
			inputDesc:      "Classic board game",
			inputCategory:  "Strategy",
			inputCondition: "good",
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				gameRepo.On("Create", mock.AnythingOfType("*models.Game")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:           "invalid game name",
			inputName:      "",
			inputDesc:      "Classic board game",
			inputCategory:  "Strategy",
			inputCondition: "good",
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "validation failed",
		},
		{
			name:           "invalid condition",
			inputName:      "Monopoly",
			inputDesc:      "Classic board game",
			inputCategory:  "Strategy",
			inputCondition: "invalid",
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "validation failed",
		},
		{
			name:           "repository error",
			inputName:      "Monopoly",
			inputDesc:      "Classic board game",
			inputCategory:  "Strategy",
			inputCondition: "good",
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				gameRepo.On("Create", mock.AnythingOfType("*models.Game")).Return(errors.New("database error"))
			},
			expectedError: "failed to create game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameRepo := &MockGameRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(gameRepo, borrowingRepo)

			service := NewGameService(gameRepo, borrowingRepo)
			game, err := service.AddGame(tt.inputName, tt.inputDesc, tt.inputCategory, tt.inputCondition)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, game)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, game)
				assert.Equal(t, tt.inputName, game.Name)
				assert.Equal(t, tt.inputDesc, game.Description)
				assert.Equal(t, tt.inputCategory, game.Category)
				assert.Equal(t, tt.inputCondition, game.Condition)
				assert.True(t, game.IsAvailable)
			}

			gameRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestGameService_GetGame(t *testing.T) {
	tests := []struct {
		name          string
		gameID        int
		setupMocks    func(*MockGameRepository, *MockBorrowingRepository)
		expectedError string
	}{
		{
			name:   "successful get game",
			gameID: 1,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				game := &models.Game{ID: 1, Name: "Monopoly", Condition: "good"}
				gameRepo.On("GetByID", 1).Return(game, nil)
			},
			expectedError: "",
		},
		{
			name:   "invalid game ID",
			gameID: 0,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid game ID",
		},
		{
			name:   "game not found",
			gameID: 999,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				gameRepo.On("GetByID", 999).Return(nil, errors.New("game not found"))
			},
			expectedError: "failed to get game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameRepo := &MockGameRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(gameRepo, borrowingRepo)

			service := NewGameService(gameRepo, borrowingRepo)
			game, err := service.GetGame(tt.gameID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, game)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, game)
				assert.Equal(t, tt.gameID, game.ID)
			}

			gameRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestGameService_SearchGames(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		setupMocks    func(*MockGameRepository, *MockBorrowingRepository)
		expectedError string
		expectedCount int
	}{
		{
			name:  "successful search with query",
			query: "Monopoly",
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				games := []*models.Game{
					{ID: 1, Name: "Monopoly", Condition: "good"},
				}
				gameRepo.On("Search", "Monopoly").Return(games, nil)
			},
			expectedError: "",
			expectedCount: 1,
		},
		{
			name:  "empty query returns all games",
			query: "",
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				games := []*models.Game{
					{ID: 1, Name: "Monopoly", Condition: "good"},
					{ID: 2, Name: "Scrabble", Condition: "excellent"},
				}
				gameRepo.On("GetAll").Return(games, nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name:  "search error",
			query: "test",
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				gameRepo.On("Search", "test").Return(nil, errors.New("search error"))
			},
			expectedError: "failed to search games",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameRepo := &MockGameRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(gameRepo, borrowingRepo)

			service := NewGameService(gameRepo, borrowingRepo)
			games, err := service.SearchGames(tt.query)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, games)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, games)
				assert.Len(t, games, tt.expectedCount)
			}

			gameRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestGameService_SetGameAvailability(t *testing.T) {
	tests := []struct {
		name          string
		gameID        int
		isAvailable   bool
		setupMocks    func(*MockGameRepository, *MockBorrowingRepository)
		expectedError string
	}{
		{
			name:        "successful availability update",
			gameID:      1,
			isAvailable: false,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				game := &models.Game{ID: 1, Name: "Monopoly", Condition: "good", IsAvailable: true}
				gameRepo.On("GetByID", 1).Return(game, nil)
				gameRepo.On("Update", mock.MatchedBy(func(g *models.Game) bool {
					return g.ID == 1 && !g.IsAvailable
				})).Return(nil)
			},
			expectedError: "",
		},
		{
			name:        "invalid game ID",
			gameID:      0,
			isAvailable: true,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid game ID",
		},
		{
			name:        "game not found",
			gameID:      999,
			isAvailable: true,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				gameRepo.On("GetByID", 999).Return(nil, errors.New("game not found"))
			},
			expectedError: "game not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameRepo := &MockGameRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(gameRepo, borrowingRepo)

			service := NewGameService(gameRepo, borrowingRepo)
			err := service.SetGameAvailability(tt.gameID, tt.isAvailable)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			gameRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestGameService_GetCurrentBorrower(t *testing.T) {
	tests := []struct {
		name          string
		gameID        int
		setupMocks    func(*MockGameRepository, *MockBorrowingRepository)
		expectedError string
		hasBorrower   bool
	}{
		{
			name:   "game has current borrower",
			gameID: 1,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
				
				borrowings := []*models.Borrowing{
					{ID: 1, UserID: 1, GameID: 1, ReturnedAt: nil}, // Active borrowing
					{ID: 2, UserID: 2, GameID: 1, ReturnedAt: &time.Time{}}, // Returned borrowing
				}
				borrowingRepo.On("GetByGame", 1).Return(borrowings, nil)
			},
			expectedError: "",
			hasBorrower:   true,
		},
		{
			name:   "game has no current borrower",
			gameID: 1,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
				
				returnTime := time.Now()
				borrowings := []*models.Borrowing{
					{ID: 1, UserID: 1, GameID: 1, ReturnedAt: &returnTime}, // Returned borrowing
				}
				borrowingRepo.On("GetByGame", 1).Return(borrowings, nil)
			},
			expectedError: "",
			hasBorrower:   false,
		},
		{
			name:   "invalid game ID",
			gameID: 0,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid game ID",
			hasBorrower:   false,
		},
		{
			name:   "game not found",
			gameID: 999,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				gameRepo.On("GetByID", 999).Return(nil, errors.New("game not found"))
			},
			expectedError: "game not found",
			hasBorrower:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameRepo := &MockGameRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(gameRepo, borrowingRepo)

			service := NewGameService(gameRepo, borrowingRepo)
			borrower, err := service.GetCurrentBorrower(tt.gameID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, borrower)
			} else {
				assert.NoError(t, err)
				if tt.hasBorrower {
					assert.NotNil(t, borrower)
					assert.Equal(t, tt.gameID, borrower.GameID)
					assert.Nil(t, borrower.ReturnedAt)
				} else {
					assert.Nil(t, borrower)
				}
			}

			gameRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}

func TestGameService_DeleteGame(t *testing.T) {
	tests := []struct {
		name          string
		gameID        int
		setupMocks    func(*MockGameRepository, *MockBorrowingRepository)
		expectedError string
	}{
		{
			name:   "successful game deletion",
			gameID: 1,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
				
				returnTime := time.Now()
				borrowings := []*models.Borrowing{
					{ID: 1, UserID: 1, GameID: 1, ReturnedAt: &returnTime}, // All returned
				}
				borrowingRepo.On("GetByGame", 1).Return(borrowings, nil)
				gameRepo.On("Delete", 1).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "cannot delete - game currently borrowed",
			gameID: 1,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				game := &models.Game{ID: 1, Name: "Monopoly"}
				gameRepo.On("GetByID", 1).Return(game, nil)
				
				borrowings := []*models.Borrowing{
					{ID: 1, UserID: 1, GameID: 1, ReturnedAt: nil}, // Currently borrowed
				}
				borrowingRepo.On("GetByGame", 1).Return(borrowings, nil)
			},
			expectedError: "cannot delete game: currently borrowed by user 1",
		},
		{
			name:   "invalid game ID",
			gameID: 0,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				// No mocks needed as validation fails first
			},
			expectedError: "invalid game ID",
		},
		{
			name:   "game not found",
			gameID: 999,
			setupMocks: func(gameRepo *MockGameRepository, borrowingRepo *MockBorrowingRepository) {
				gameRepo.On("GetByID", 999).Return(nil, errors.New("game not found"))
			},
			expectedError: "game not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameRepo := &MockGameRepository{}
			borrowingRepo := &MockBorrowingRepository{}
			tt.setupMocks(gameRepo, borrowingRepo)

			service := NewGameService(gameRepo, borrowingRepo)
			err := service.DeleteGame(tt.gameID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			gameRepo.AssertExpectations(t)
			borrowingRepo.AssertExpectations(t)
		})
	}
}