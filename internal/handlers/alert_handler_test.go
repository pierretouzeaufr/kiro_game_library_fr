package handlers

import (
	"board-game-library/internal/models"
	"board-game-library/internal/services"
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

// MockAlertService is a mock implementation of AlertServiceInterface
type MockAlertService struct {
	mock.Mock
}

func (m *MockAlertService) GenerateOverdueAlerts() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAlertService) GenerateReminderAlerts() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAlertService) GetActiveAlerts() ([]*models.Alert, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Alert), args.Error(1)
}

func (m *MockAlertService) GetAlertsByUser(userID int) ([]*models.Alert, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Alert), args.Error(1)
}

func (m *MockAlertService) MarkAlertAsRead(alertID int) error {
	args := m.Called(alertID)
	return args.Error(0)
}

func (m *MockAlertService) MarkAllUserAlertsAsRead(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAlertService) DeleteAlert(alertID int) error {
	args := m.Called(alertID)
	return args.Error(0)
}

func (m *MockAlertService) CleanupResolvedAlerts() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAlertService) GetAlertsSummaryByUser() (map[int]services.AlertSummary, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]services.AlertSummary), args.Error(1)
}

func (m *MockAlertService) CreateCustomAlert(userID, gameID int, alertType, message string) (*models.Alert, error) {
	args := m.Called(userID, gameID, alertType, message)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Alert), args.Error(1)
}

func setupAlertHandlerTest() (*gin.Engine, *MockAlertService, *AlertHandler) {
	gin.SetMode(gin.TestMode)
	
	mockService := &MockAlertService{}
	handler := &AlertHandler{
		alertService: mockService,
	}
	
	router := gin.New()
	api := router.Group("/api")
	handler.RegisterRoutes(api)
	
	return router, mockService, handler
}

func TestAlertHandler_GetActiveAlerts(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedAlerts := []*models.Alert{
			{ID: 1, UserID: 1, GameID: 1, Type: "overdue", IsRead: false},
			{ID: 2, UserID: 2, GameID: 2, Type: "reminder", IsRead: false},
		}

		mockService.On("GetActiveAlerts").Return(expectedAlerts, nil)

		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.NotNil(t, response["alerts"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupAlertHandlerTest()
		mockService.On("GetActiveAlerts").Return(nil, fmt.Errorf("database error"))

		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_GetAlertsByUser(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedAlerts := []*models.Alert{
			{ID: 1, UserID: 1, GameID: 1, Type: "overdue", IsRead: false},
			{ID: 2, UserID: 1, GameID: 2, Type: "reminder", IsRead: true},
		}

		mockService.On("GetAlertsByUser", 1).Return(expectedAlerts, nil)

		req, _ := http.NewRequest("GET", "/api/alerts/user/1", nil)
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

	t.Run("filter unread only", func(t *testing.T) {
		expectedAlerts := []*models.Alert{
			{ID: 1, UserID: 1, GameID: 1, Type: "overdue", IsRead: false},
			{ID: 2, UserID: 1, GameID: 2, Type: "reminder", IsRead: true},
		}

		mockService.On("GetAlertsByUser", 1).Return(expectedAlerts, nil)

		req, _ := http.NewRequest("GET", "/api/alerts/user/1?unread=true", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"]) // Only unread alerts

		mockService.AssertExpectations(t)
	})

	t.Run("filter by type", func(t *testing.T) {
		expectedAlerts := []*models.Alert{
			{ID: 1, UserID: 1, GameID: 1, Type: "overdue", IsRead: false},
			{ID: 2, UserID: 1, GameID: 2, Type: "reminder", IsRead: false},
		}

		mockService.On("GetAlertsByUser", 1).Return(expectedAlerts, nil)

		req, _ := http.NewRequest("GET", "/api/alerts/user/1?type=overdue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"]) // Only overdue alerts

		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("GetAlertsByUser", 999).Return(nil, fmt.Errorf("user not found"))

		req, _ := http.NewRequest("GET", "/api/alerts/user/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/user/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAlertHandler_MarkAlertAsRead(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful mark as read", func(t *testing.T) {
		mockService.On("MarkAlertAsRead", 1).Return(nil)

		req, _ := http.NewRequest("PUT", "/api/alerts/1/read", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Alert marked as read successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("alert not found", func(t *testing.T) {
		mockService.On("MarkAlertAsRead", 999).Return(fmt.Errorf("alert not found"))

		req, _ := http.NewRequest("PUT", "/api/alerts/999/read", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid alert ID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/alerts/invalid/read", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAlertHandler_MarkAllUserAlertsAsRead(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful mark all as read", func(t *testing.T) {
		mockService.On("MarkAllUserAlertsAsRead", 1).Return(nil)

		req, _ := http.NewRequest("PUT", "/api/alerts/user/1/read-all", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "All user alerts marked as read successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("MarkAllUserAlertsAsRead", 999).Return(fmt.Errorf("user not found"))

		req, _ := http.NewRequest("PUT", "/api/alerts/user/999/read-all", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_DeleteAlert(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful deletion", func(t *testing.T) {
		mockService.On("DeleteAlert", 1).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/alerts/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Alert deleted successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("alert not found", func(t *testing.T) {
		mockService.On("DeleteAlert", 999).Return(fmt.Errorf("alert not found"))

		req, _ := http.NewRequest("DELETE", "/api/alerts/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_GetAlertsSummary(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedSummary := map[int]services.AlertSummary{
			1: {
				UserID:        1,
				TotalAlerts:   2,
				OverdueCount:  1,
				ReminderCount: 1,
			},
			2: {
				UserID:        2,
				TotalAlerts:   1,
				OverdueCount:  1,
				ReminderCount: 0,
			},
		}

		mockService.On("GetAlertsSummaryByUser").Return(expectedSummary, nil)

		req, _ := http.NewRequest("GET", "/api/alerts/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["users"])
		assert.NotNil(t, response["summary"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupAlertHandlerTest()
		mockService.On("GetAlertsSummaryByUser").Return(nil, fmt.Errorf("database error"))

		req, _ := http.NewRequest("GET", "/api/alerts/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_GetDashboard(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful retrieval", func(t *testing.T) {
		expectedSummary := map[int]services.AlertSummary{
			1: {
				UserID:        1,
				TotalAlerts:   2,
				OverdueCount:  1,
				ReminderCount: 1,
			},
			2: {
				UserID:        2,
				TotalAlerts:   1,
				OverdueCount:  1,
				ReminderCount: 0,
			},
		}

		mockService.On("GetAlertsSummaryByUser").Return(expectedSummary, nil)

		req, _ := http.NewRequest("GET", "/api/alerts/dashboard", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		dashboard := response["dashboard"].(map[string]interface{})
		assert.Equal(t, float64(3), dashboard["total_alerts"])      // 2 + 1
		assert.Equal(t, float64(2), dashboard["total_overdue"])     // 1 + 1
		assert.Equal(t, float64(1), dashboard["total_reminders"])   // 1 + 0
		assert.Equal(t, float64(2), dashboard["users_with_alerts"]) // 2 users

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_CreateCustomAlert(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful creation", func(t *testing.T) {
		expectedAlert := &models.Alert{
			ID:        1,
			UserID:    1,
			GameID:    1,
			Type:      "custom",
			Message:   "Custom alert message",
			CreatedAt: time.Now(),
			IsRead:    false,
		}

		mockService.On("CreateCustomAlert", 1, 1, "custom", "Custom alert message").Return(expectedAlert, nil)

		reqBody := CreateCustomAlertRequest{
			UserID:  1,
			GameID:  1,
			Type:    "custom",
			Message: "Custom alert message",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/alerts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Alert created successfully", response["message"])
		assert.NotNil(t, response["alert"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request data", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"user_id": 1,
			// Missing required fields
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/alerts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("CreateCustomAlert", 999, 1, "custom", "Custom alert message").Return(nil, fmt.Errorf("user not found"))

		reqBody := CreateCustomAlertRequest{
			UserID:  999,
			GameID:  1,
			Type:    "custom",
			Message: "Custom alert message",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/alerts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_GenerateOverdueAlerts(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful generation", func(t *testing.T) {
		mockService.On("GenerateOverdueAlerts").Return(nil)

		req, _ := http.NewRequest("POST", "/api/alerts/generate-overdue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Overdue alerts generated successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupAlertHandlerTest()
		mockService.On("GenerateOverdueAlerts").Return(fmt.Errorf("database error"))

		req, _ := http.NewRequest("POST", "/api/alerts/generate-overdue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_GenerateReminderAlerts(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful generation", func(t *testing.T) {
		mockService.On("GenerateReminderAlerts").Return(nil)

		req, _ := http.NewRequest("POST", "/api/alerts/generate-reminders", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Reminder alerts generated successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupAlertHandlerTest()
		mockService.On("GenerateReminderAlerts").Return(fmt.Errorf("database error"))

		req, _ := http.NewRequest("POST", "/api/alerts/generate-reminders", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAlertHandler_CleanupResolvedAlerts(t *testing.T) {
	router, mockService, _ := setupAlertHandlerTest()

	t.Run("successful cleanup", func(t *testing.T) {
		mockService.On("CleanupResolvedAlerts").Return(nil)

		req, _ := http.NewRequest("POST", "/api/alerts/cleanup", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Resolved alerts cleaned up successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		router, mockService, _ := setupAlertHandlerTest()
		mockService.On("CleanupResolvedAlerts").Return(fmt.Errorf("database error"))

		req, _ := http.NewRequest("POST", "/api/alerts/cleanup", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}