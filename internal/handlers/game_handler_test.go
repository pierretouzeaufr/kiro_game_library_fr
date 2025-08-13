package handlers

import (
	"board-game-library/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGameService is a mock implementation of GameServiceInterface
type MockGameService struct {
	mock.Mock
}

func (m *MockGameService) AddGame(name, description, category, condition string) (*models.Game, error) {
	args := m.Called(name, description, category, condition)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Game), args.Error(1)
}

func (m *MockGameService) GetGame(id int) (*models.Game, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Game), args.Error(1)
}

func (m *MockGameService) GetAllGames() ([]*models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameService) GetAvailableGames() ([]*models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameService) SearchGames(query string) ([]*models.Game, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Game), args.Error(1)
}

func (m *MockGameService) UpdateGame(game *models.Game) error {
	args := m.Called(game)
	return args.Error(0)
}

func (m *MockGameService) SetGameAvailability(gameID int, isAvailable bool) error {
	args := m.Called(gameID, isAvailable)
	return args.Error(0)
}

func (m *MockGameService) GetGameBorrowingHistory(gameID int) ([]*models.Borrowing, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockGameService) IsGameAvailable(gameID int) (bool, error) {
	args := m.Called(gameID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGameService) GetCurrentBorrower(gameID int) (*models.Borrowing, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Borrowing), args.Error(1)
}

func (m *MockGameService) DeleteGame(gameID int) error {
	args := m.Called(gameID)
	return args.Error(0)
}

func setupGameHandlerTest() (*gin.Engine, *MockGameService, *GameHandler) {
	gin.SetMode(gin.TestMode)
	
	mockService := &MockGameService{}
	handler := &GameHandler{
		gameService: mockService,
	}
	
	router := gin.New()
	api := router.Group("/api")
	handler.RegisterRoutes(api)
	
	return router, mockService, handler
}

func TestGameHandler_AddGame(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("successful game addition", func(t *testing.T) {
		expectedGame := &models.Game{
			ID:          1,
			Name:        "Monopoly",
			Description: "Classic board game",
			Category:    "Strategy",
			EntryDate:   time.Now(),
			Condition:   "good",
			IsAvailable: true,
		}

		mockService.On("AddGame", "Monopoly", "Classic board game", "Strategy", "good").Return(expectedGame, nil)

		reqBody := AddGameRequest{
			Name:        "Monopoly",
			Description: "Classic board game",
			Category:    "Strategy",
			Condition:   "good",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Game added successfully", response["message"])
		assert.NotNil(t, response["game"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request data", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"description": "Missing name",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request data", response["error"])
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupGameHandlerTest()
		mockService.On("AddGame", "Monopoly", "Classic board game", "Strategy", "good").Return(nil, fmt.Errorf("database error"))

		reqBody := AddGameRequest{
			Name:        "Monopoly",
			Description: "Classic board game",
			Category:    "Strategy",
			Condition:   "good",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGameHandler_GetAllGames(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("get all games", func(t *testing.T) {
		expectedGames := []*models.Game{
			{ID: 1, Name: "Monopoly", IsAvailable: true},
			{ID: 2, Name: "Scrabble", IsAvailable: false},
		}

		mockService.On("GetAllGames").Return(expectedGames, nil)

		req, _ := http.NewRequest("GET", "/api/games", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.NotNil(t, response["games"])

		mockService.AssertExpectations(t)
	})

	t.Run("get available games only", func(t *testing.T) {
		expectedGames := []*models.Game{
			{ID: 1, Name: "Monopoly", IsAvailable: true},
		}

		mockService.On("GetAvailableGames").Return(expectedGames, nil)

		req, _ := http.NewRequest("GET", "/api/games?available=true", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"])

		mockService.AssertExpectations(t)
	})

	t.Run("search games", func(t *testing.T) {
		expectedGames := []*models.Game{
			{ID: 1, Name: "Monopoly", IsAvailable: true},
		}

		mockService.On("SearchGames", "Monopoly").Return(expectedGames, nil)

		req, _ := http.NewRequest("GET", "/api/games?search=Monopoly", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGameHandler_GetGame(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedGame := &models.Game{
			ID:          1,
			Name:        "Monopoly",
			Description: "Classic board game",
			IsAvailable: true,
		}

		mockService.On("GetGame", 1).Return(expectedGame, nil)

		req, _ := http.NewRequest("GET", "/api/games/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["game"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid game ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/games/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("game not found", func(t *testing.T) {
		mockService.On("GetGame", 999).Return(nil, fmt.Errorf("game not found"))

		req, _ := http.NewRequest("GET", "/api/games/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGameHandler_UpdateGame(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("successful update", func(t *testing.T) {
		existingGame := &models.Game{
			ID:          1,
			Name:        "Monopoly",
			Description: "Classic board game",
			Category:    "Strategy",
			Condition:   "good",
			IsAvailable: true,
		}

		isAvailable := false
		mockService.On("GetGame", 1).Return(existingGame, nil)
		mockService.On("UpdateGame", mock.MatchedBy(func(game *models.Game) bool {
			return game.ID == 1 && game.Name == "Monopoly Deluxe" && game.IsAvailable == false
		})).Return(nil)

		reqBody := UpdateGameRequest{
			Name:        "Monopoly Deluxe",
			Description: "Deluxe edition",
			Category:    "Strategy",
			Condition:   "excellent",
			IsAvailable: &isAvailable,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/games/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("game not found", func(t *testing.T) {
		mockService.On("GetGame", 999).Return(nil, fmt.Errorf("game not found"))

		reqBody := UpdateGameRequest{
			Name:        "Monopoly Deluxe",
			Description: "Deluxe edition",
			Category:    "Strategy",
			Condition:   "excellent",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/games/999", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGameHandler_DeleteGame(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("successful deletion", func(t *testing.T) {
		mockService.On("DeleteGame", 1).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/games/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Game deleted successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("game not found", func(t *testing.T) {
		mockService.On("DeleteGame", 999).Return(fmt.Errorf("game not found"))

		req, _ := http.NewRequest("DELETE", "/api/games/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("game currently borrowed", func(t *testing.T) {
		router, mockService, _ := setupGameHandlerTest()
		mockService.On("DeleteGame", 1).Return(fmt.Errorf("cannot delete game: currently borrowed"))

		req, _ := http.NewRequest("DELETE", "/api/games/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGameHandler_GetGameBorrowingHistory(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedBorrowings := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1, BorrowedAt: time.Now()},
			{ID: 2, UserID: 2, GameID: 1, BorrowedAt: time.Now().AddDate(0, 0, -30)},
		}

		mockService.On("GetGameBorrowingHistory", 1).Return(expectedBorrowings, nil)

		req, _ := http.NewRequest("GET", "/api/games/1/borrowings", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])

		mockService.AssertExpectations(t)
	})

	t.Run("game not found", func(t *testing.T) {
		mockService.On("GetGameBorrowingHistory", 999).Return(nil, fmt.Errorf("game not found"))

		req, _ := http.NewRequest("GET", "/api/games/999/borrowings", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGameHandler_GetGameAvailability(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("game is available", func(t *testing.T) {
		mockService.On("IsGameAvailable", 1).Return(true, nil)

		req, _ := http.NewRequest("GET", "/api/games/1/availability", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["is_available"])

		mockService.AssertExpectations(t)
	})

	t.Run("game is not available", func(t *testing.T) {
		router, mockService, _ := setupGameHandlerTest()
		currentBorrower := &models.Borrowing{
			ID:     1,
			UserID: 1,
			GameID: 1,
		}

		mockService.On("IsGameAvailable", 1).Return(false, nil)
		mockService.On("GetCurrentBorrower", 1).Return(currentBorrower, nil)

		req, _ := http.NewRequest("GET", "/api/games/1/availability", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, false, response["is_available"])
		assert.NotNil(t, response["current_borrower"])

		mockService.AssertExpectations(t)
	})
}

func TestGameHandler_SearchGames(t *testing.T) {
	router, mockService, _ := setupGameHandlerTest()

	t.Run("successful search", func(t *testing.T) {
		expectedGames := []*models.Game{
			{ID: 1, Name: "Monopoly", IsAvailable: true},
		}

		mockService.On("SearchGames", "Monopoly").Return(expectedGames, nil)

		req, _ := http.NewRequest("GET", "/api/games/search?q=Monopoly", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"])
		assert.Equal(t, "Monopoly", response["query"])

		mockService.AssertExpectations(t)
	})

	t.Run("missing search query", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/games/search", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Missing search query", response["error"])
	})
}