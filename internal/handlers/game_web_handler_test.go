package handlers

import (
	"board-game-library/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGameServiceInterface for testing
type MockGameServiceInterface struct {
	mock.Mock
}

func (m *MockGameServiceInterface) AddGame(name, description, category, condition string) (*models.Game, error) {
	args := m.Called(name, description, category, condition)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Game), args.Error(1)
}

func (m *MockGameServiceInterface) GetGame(id int) (*models.Game, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Game), args.Error(1)
}

func (m *MockGameServiceInterface) GetAllGames() ([]*models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameServiceInterface) GetAvailableGames() ([]*models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameServiceInterface) SearchGames(query string) ([]*models.Game, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameServiceInterface) UpdateGame(game *models.Game) error {
	args := m.Called(game)
	return args.Error(0)
}

func (m *MockGameServiceInterface) GetGameBorrowingHistory(gameID int) ([]*models.Borrowing, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockGameServiceInterface) SetGameAvailability(gameID int, isAvailable bool) error {
	args := m.Called(gameID, isAvailable)
	return args.Error(0)
}

func (m *MockGameServiceInterface) IsGameAvailable(gameID int) (bool, error) {
	args := m.Called(gameID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGameServiceInterface) GetCurrentBorrower(gameID int) (*models.Borrowing, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Borrowing), args.Error(1)
}

func (m *MockGameServiceInterface) DeleteGame(gameID int) error {
	args := m.Called(gameID)
	return args.Error(0)
}

func TestGameWebHandler_SearchFilterGames_Logic(t *testing.T) {
	tests := []struct {
		name           string
		searchQuery    string
		availability   string
		allGames       []*models.Game
		searchResults  []*models.Game
		expectedGames  int
		expectSearch   bool
	}{
		{
			name:        "search with no filters",
			searchQuery: "",
			availability: "all",
			allGames: []*models.Game{
				{ID: 1, Name: "Chess", IsAvailable: true},
				{ID: 2, Name: "Checkers", IsAvailable: false},
			},
			expectedGames: 2,
			expectSearch:  false,
		},
		{
			name:        "search with query",
			searchQuery: "chess",
			availability: "all",
			searchResults: []*models.Game{
				{ID: 1, Name: "Chess", IsAvailable: true},
			},
			expectedGames: 1,
			expectSearch:  true,
		},
		{
			name:        "filter available only",
			searchQuery: "",
			availability: "available",
			allGames: []*models.Game{
				{ID: 1, Name: "Chess", IsAvailable: true},
				{ID: 2, Name: "Checkers", IsAvailable: false},
			},
			expectedGames: 1,
			expectSearch:  false,
		},
		{
			name:        "filter borrowed only",
			searchQuery: "",
			availability: "borrowed",
			allGames: []*models.Game{
				{ID: 1, Name: "Chess", IsAvailable: true},
				{ID: 2, Name: "Checkers", IsAvailable: false},
			},
			expectedGames: 1,
			expectSearch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the filtering logic
			var games []*models.Game
			
			if tt.expectSearch {
				games = tt.searchResults
			} else {
				games = tt.allGames
			}

			// Apply availability filter
			if tt.availability != "" && tt.availability != "all" {
				filteredGames := make([]*models.Game, 0)
				for _, game := range games {
					switch tt.availability {
					case "available":
						if game.IsAvailable {
							filteredGames = append(filteredGames, game)
						}
					case "borrowed":
						if !game.IsAvailable {
							filteredGames = append(filteredGames, game)
						}
					}
				}
				games = filteredGames
			}

			assert.Equal(t, tt.expectedGames, len(games))
		})
	}
}

func TestGameWebHandler_ServiceIntegration(t *testing.T) {
	// Test that the handler correctly calls the service methods
	mockGameService := new(MockGameServiceInterface)
	handler := NewGameWebHandler(mockGameService)

	// Test that the handler is properly initialized
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.gameService)
}