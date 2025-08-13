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

// MockUserService is a mock implementation of UserServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(name, email string) (*models.User, error) {
	args := m.Called(name, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUser(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) GetUserBorrowings(userID int) ([]*models.Borrowing, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockUserService) CanUserBorrow(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) GetActiveUserBorrowings(userID int) ([]*models.Borrowing, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Borrowing), args.Error(1)
}

func (m *MockUserService) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func setupUserHandlerTest() (*gin.Engine, *MockUserService, *UserHandler) {
	gin.SetMode(gin.TestMode)
	
	mockService := &MockUserService{}
	handler := &UserHandler{
		userService: mockService,
	}
	
	router := gin.New()
	api := router.Group("/api")
	handler.RegisterRoutes(api)
	
	return router, mockService, handler
}

func TestUserHandler_RegisterUser(t *testing.T) {
	router, mockService, _ := setupUserHandlerTest()

	t.Run("successful registration", func(t *testing.T) {
		expectedUser := &models.User{
			ID:           1,
			Name:         "John Doe",
			Email:        "john@example.com",
			RegisteredAt: time.Now(),
			IsActive:     true,
		}

		mockService.On("RegisterUser", "John Doe", "john@example.com").Return(expectedUser, nil)

		reqBody := RegisterUserRequest{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User registered successfully", response["message"])
		assert.NotNil(t, response["user"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request data", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "John Doe",
			// Missing email
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request data", response["error"])
	})

	t.Run("user already exists", func(t *testing.T) {
		mockService.On("RegisterUser", "Jane Doe", "jane@example.com").Return(nil, fmt.Errorf("user with email jane@example.com already exists"))

		reqBody := RegisterUserRequest{
			Name:  "Jane Doe",
			Email: "jane@example.com",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User already exists", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	router, mockService, _ := setupUserHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedUsers := []*models.User{
			{ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true},
			{ID: 2, Name: "Jane Doe", Email: "jane@example.com", IsActive: true},
		}

		mockService.On("GetAllUsers").Return(expectedUsers, nil)

		req, _ := http.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.NotNil(t, response["users"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupUserHandlerTest()
		mockService.On("GetAllUsers").Return(nil, fmt.Errorf("database error"))

		req, _ := http.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve users", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUser(t *testing.T) {
	router, mockService, _ := setupUserHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedUser := &models.User{
			ID:       1,
			Name:     "John Doe",
			Email:    "john@example.com",
			IsActive: true,
		}

		mockService.On("GetUser", 1).Return(expectedUser, nil)

		req, _ := http.NewRequest("GET", "/api/users/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["user"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid user ID", response["error"])
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("GetUser", 999).Return(nil, fmt.Errorf("user not found"))

		req, _ := http.NewRequest("GET", "/api/users/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not found", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUserBorrowings(t *testing.T) {
	router, mockService, _ := setupUserHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedBorrowings := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1, BorrowedAt: time.Now(), DueDate: time.Now().AddDate(0, 0, 14)},
			{ID: 2, UserID: 1, GameID: 2, BorrowedAt: time.Now().AddDate(0, 0, -30), DueDate: time.Now().AddDate(0, 0, -16)},
		}

		mockService.On("GetUserBorrowings", 1).Return(expectedBorrowings, nil)

		req, _ := http.NewRequest("GET", "/api/users/1/borrowings", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.NotNil(t, response["borrowings"])

		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("GetUserBorrowings", 999).Return(nil, fmt.Errorf("user not found"))

		req, _ := http.NewRequest("GET", "/api/users/999/borrowings", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUserCurrentLoans(t *testing.T) {
	router, mockService, _ := setupUserHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedLoans := []*models.Borrowing{
			{ID: 1, UserID: 1, GameID: 1, BorrowedAt: time.Now(), DueDate: time.Now().AddDate(0, 0, 14)},
		}

		mockService.On("GetActiveUserBorrowings", 1).Return(expectedLoans, nil)

		req, _ := http.NewRequest("GET", "/api/users/1/current-loans", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"])
		assert.NotNil(t, response["current_loans"])

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_CheckUserEligibility(t *testing.T) {
	router, mockService, _ := setupUserHandlerTest()

	t.Run("user can borrow", func(t *testing.T) {
		mockService.On("CanUserBorrow", 1).Return(true, nil)

		req, _ := http.NewRequest("GET", "/api/users/1/eligibility", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["can_borrow"])

		mockService.AssertExpectations(t)
	})

	t.Run("user has overdue items", func(t *testing.T) {
		router, mockService, _ := setupUserHandlerTest()
		mockService.On("CanUserBorrow", 1).Return(false, fmt.Errorf("user has overdue items"))

		req, _ := http.NewRequest("GET", "/api/users/1/eligibility", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, false, response["can_borrow"])
		assert.Equal(t, "user has overdue items", response["reason"])

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	router, mockService, _ := setupUserHandlerTest()

	t.Run("successful update", func(t *testing.T) {
		existingUser := &models.User{
			ID:       1,
			Name:     "John Doe",
			Email:    "john@example.com",
			IsActive: true,
		}

		isActive := false
		mockService.On("GetUser", 1).Return(existingUser, nil)
		mockService.On("UpdateUser", mock.MatchedBy(func(user *models.User) bool {
			return user.ID == 1 && user.Name == "John Smith" && user.Email == "johnsmith@example.com" && user.IsActive == false
		})).Return(nil)

		reqBody := UpdateUserRequest{
			Name:     "John Smith",
			Email:    "johnsmith@example.com",
			IsActive: &isActive,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/users/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User updated successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("GetUser", 999).Return(nil, fmt.Errorf("user not found"))

		reqBody := UpdateUserRequest{
			Name:  "John Smith",
			Email: "johnsmith@example.com",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/users/999", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}