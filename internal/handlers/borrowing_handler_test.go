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

// MockBorrowingService is a mock implementation of BorrowingServiceInterface
type MockBorrowingService struct {
	mock.Mock
}

func (m *MockBorrowingService) BorrowGame(userID, gameID int, dueDate time.Time) (*models.Borrowing, error) {
	args := m.Called(userID, gameID, dueDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingService) BorrowGameWithDefaultDueDate(userID, gameID int) (*models.Borrowing, error) {
	args := m.Called(userID, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingService) ReturnGame(borrowingID int) error {
	args := m.Called(borrowingID)
	return args.Error(0)
}

func (m *MockBorrowingService) GetOverdueItems() ([]*models.Borrowing, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingService) ExtendDueDate(borrowingID int, newDueDate time.Time) error {
	args := m.Called(borrowingID, newDueDate)
	return args.Error(0)
}

func (m *MockBorrowingService) GetBorrowingDetails(borrowingID int) (*models.Borrowing, error) {
	args := m.Called(borrowingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingService) GetActiveBorrowingsByUser(userID int) ([]*models.Borrowing, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingService) GetBorrowingsByGame(gameID int) ([]*models.Borrowing, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockBorrowingService) UpdateOverdueStatus() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBorrowingService) GetItemsDueSoon(daysAhead int) ([]*models.Borrowing, error) {
	args := m.Called(daysAhead)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func setupBorrowingHandlerTest() (*gin.Engine, *MockBorrowingService, *BorrowingHandler) {
	gin.SetMode(gin.TestMode)
	
	mockService := &MockBorrowingService{}
	handler := &BorrowingHandler{
		borrowingService: mockService,
	}
	
	router := gin.New()
	api := router.Group("/api")
	handler.RegisterRoutes(api)
	
	return router, mockService, handler
}

func TestBorrowingHandler_BorrowGame(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful borrowing with default due date", func(t *testing.T) {
		expectedBorrowing := &models.Borrowing{
			ID:         1,
			UserID:     1,
			GameID:     1,
			BorrowedAt: time.Now(),
			DueDate:    time.Now().Add(14 * 24 * time.Hour),
		}

		mockService.On("BorrowGameWithDefaultDueDate", 1, 1).Return(expectedBorrowing, nil)

		reqBody := BorrowGameRequest{
			UserID: 1,
			GameID: 1,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/borrowings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Game borrowed successfully", response["message"])
		assert.NotNil(t, response["borrowing"])

		mockService.AssertExpectations(t)
	})

	t.Run("successful borrowing with custom due date", func(t *testing.T) {
		dueDate := time.Now().Add(7 * 24 * time.Hour)
		expectedBorrowing := &models.Borrowing{
			ID:         1,
			UserID:     1,
			GameID:     1,
			BorrowedAt: time.Now(),
			DueDate:    dueDate,
		}

		mockService.On("BorrowGame", 1, 1, mock.MatchedBy(func(t time.Time) bool {
			return t.Format("2006-01-02") == dueDate.Format("2006-01-02")
		})).Return(expectedBorrowing, nil)

		reqBody := BorrowGameRequest{
			UserID:  1,
			GameID:  1,
			DueDate: dueDate.Format("2006-01-02"),
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/borrowings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request data", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"user_id": 1,
			// Missing game_id
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/borrowings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("game not available", func(t *testing.T) {
		router, mockService, _ := setupBorrowingHandlerTest()
		mockService.On("BorrowGameWithDefaultDueDate", 1, 1).Return(nil, fmt.Errorf("game is not available for borrowing"))

		reqBody := BorrowGameRequest{
			UserID: 1,
			GameID: 1,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/borrowings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("BorrowGameWithDefaultDueDate", 999, 1).Return(nil, fmt.Errorf("user not found"))

		reqBody := BorrowGameRequest{
			UserID: 999,
			GameID: 1,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/borrowings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestBorrowingHandler_ReturnGame(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful return", func(t *testing.T) {
		mockService.On("ReturnGame", 1).Return(nil)

		req, _ := http.NewRequest("PUT", "/api/borrowings/1/return", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Game returned successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid borrowing ID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/borrowings/invalid/return", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("borrowing not found", func(t *testing.T) {
		mockService.On("ReturnGame", 999).Return(fmt.Errorf("borrowing not found"))

		req, _ := http.NewRequest("PUT", "/api/borrowings/999/return", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("already returned", func(t *testing.T) {
		router, mockService, _ := setupBorrowingHandlerTest()
		mockService.On("ReturnGame", 1).Return(fmt.Errorf("game has already been returned"))

		req, _ := http.NewRequest("PUT", "/api/borrowings/1/return", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestBorrowingHandler_GetBorrowingDetails(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedBorrowing := &models.Borrowing{
			ID:         1,
			UserID:     1,
			GameID:     1,
			BorrowedAt: time.Now(),
			DueDate:    time.Now().Add(14 * 24 * time.Hour),
		}

		mockService.On("GetBorrowingDetails", 1).Return(expectedBorrowing, nil)

		req, _ := http.NewRequest("GET", "/api/borrowings/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["borrowing"])

		mockService.AssertExpectations(t)
	})

	t.Run("borrowing not found", func(t *testing.T) {
		mockService.On("GetBorrowingDetails", 999).Return(nil, fmt.Errorf("borrowing not found"))

		req, _ := http.NewRequest("GET", "/api/borrowings/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestBorrowingHandler_ExtendDueDate(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful extension", func(t *testing.T) {
		newDueDate := time.Now().Add(21 * 24 * time.Hour)
		mockService.On("ExtendDueDate", 1, mock.MatchedBy(func(t time.Time) bool {
			return t.Format("2006-01-02") == newDueDate.Format("2006-01-02")
		})).Return(nil)

		reqBody := ExtendDueDateRequest{
			NewDueDate: newDueDate.Format("2006-01-02"),
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/borrowings/1/extend", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Due date extended successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid date format", func(t *testing.T) {
		reqBody := ExtendDueDateRequest{
			NewDueDate: "invalid-date",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/borrowings/1/extend", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("borrowing not found", func(t *testing.T) {
		newDueDate := time.Now().Add(21 * 24 * time.Hour)
		mockService.On("ExtendDueDate", 999, mock.AnythingOfType("time.Time")).Return(fmt.Errorf("borrowing not found"))

		reqBody := ExtendDueDateRequest{
			NewDueDate: newDueDate.Format("2006-01-02"),
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/borrowings/999/extend", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestBorrowingHandler_GetOverdueItems(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedItems := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1, IsOverdue: true},
			{ID: 2, UserID: 2, GameID: 2, IsOverdue: true},
		}

		mockService.On("GetOverdueItems").Return(expectedItems, nil)

		req, _ := http.NewRequest("GET", "/api/borrowings/overdue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.NotNil(t, response["overdue_items"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupBorrowingHandlerTest()
		mockService.On("GetOverdueItems").Return(nil, fmt.Errorf("database error"))

		req, _ := http.NewRequest("GET", "/api/borrowings/overdue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestBorrowingHandler_GetItemsDueSoon(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful retrieval with default days", func(t *testing.T) {
		expectedItems := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1, DueDate: time.Now().Add(24 * time.Hour)},
		}

		mockService.On("GetItemsDueSoon", 2).Return(expectedItems, nil)

		req, _ := http.NewRequest("GET", "/api/borrowings/due-soon", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"])
		assert.Equal(t, float64(2), response["days_ahead"])

		mockService.AssertExpectations(t)
	})

	t.Run("successful retrieval with custom days", func(t *testing.T) {
		expectedItems := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1, DueDate: time.Now().Add(3 * 24 * time.Hour)},
		}

		mockService.On("GetItemsDueSoon", 5).Return(expectedItems, nil)

		req, _ := http.NewRequest("GET", "/api/borrowings/due-soon?days=5", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(5), response["days_ahead"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid days parameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/borrowings/due-soon?days=invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBorrowingHandler_GetActiveBorrowingsByUser(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedBorrowings := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1},
			{ID: 2, UserID: 1, GameID: 2},
		}

		mockService.On("GetActiveBorrowingsByUser", 1).Return(expectedBorrowings, nil)

		req, _ := http.NewRequest("GET", "/api/borrowings/user/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.Equal(t, float64(1), response["user_id"])

		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("GetActiveBorrowingsByUser", 999).Return(nil, fmt.Errorf("user not found"))

		req, _ := http.NewRequest("GET", "/api/borrowings/user/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestBorrowingHandler_GetBorrowingsByGame(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedBorrowings := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1},
			{ID: 2, UserID: 2, GameID: 1},
		}

		mockService.On("GetBorrowingsByGame", 1).Return(expectedBorrowings, nil)

		req, _ := http.NewRequest("GET", "/api/borrowings/game/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.Equal(t, float64(1), response["game_id"])

		mockService.AssertExpectations(t)
	})

	t.Run("game not found", func(t *testing.T) {
		mockService.On("GetBorrowingsByGame", 999).Return(nil, fmt.Errorf("game not found"))

		req, _ := http.NewRequest("GET", "/api/borrowings/game/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestBorrowingHandler_UpdateOverdueStatus(t *testing.T) {
	router, mockService, _ := setupBorrowingHandlerTest()

	t.Run("successful update", func(t *testing.T) {
		mockService.On("UpdateOverdueStatus").Return(nil)

		req, _ := http.NewRequest("POST", "/api/borrowings/update-overdue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Overdue status updated successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupBorrowingHandlerTest()
		mockService.On("UpdateOverdueStatus").Return(fmt.Errorf("database error"))

		req, _ := http.NewRequest("POST", "/api/borrowings/update-overdue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}