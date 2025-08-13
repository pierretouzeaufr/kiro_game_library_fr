package handlers

import (
	"board-game-library/internal/models"
	"board-game-library/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AlertServiceInterface defines the interface for alert service operations
type AlertServiceInterface interface {
	GenerateOverdueAlerts() error
	GenerateReminderAlerts() error
	GetActiveAlerts() ([]*models.Alert, error)
	GetAlertsByUser(userID int) ([]*models.Alert, error)
	MarkAlertAsRead(alertID int) error
	MarkAllUserAlertsAsRead(userID int) error
	DeleteAlert(alertID int) error
	CleanupResolvedAlerts() error
	GetAlertsSummaryByUser() (map[int]services.AlertSummary, error)
	CreateCustomAlert(userID, gameID int, alertType, message string) (*models.Alert, error)
}

// AlertHandler handles HTTP requests for alert management
type AlertHandler struct {
	alertService AlertServiceInterface
}

// NewAlertHandler creates a new AlertHandler instance
func NewAlertHandler(alertService AlertServiceInterface) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// GetActiveAlerts handles GET /api/alerts - get all unread alerts
func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve active alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetAlertsByUser handles GET /api/alerts/user/:id - get alerts for a specific user
func (h *AlertHandler) GetAlertsByUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	// Check for filter parameters
	unreadOnly := c.Query("unread") == "true"
	alertType := c.Query("type")

	alerts, err := h.alertService.GetAlertsByUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user alerts",
			"details": err.Error(),
		})
		return
	}

	// Apply filters
	var filteredAlerts []*models.Alert
	for _, alert := range alerts {
		// Filter by read status
		if unreadOnly && alert.IsRead {
			continue
		}
		// Filter by type
		if alertType != "" && alert.Type != alertType {
			continue
		}
		filteredAlerts = append(filteredAlerts, alert)
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts":  filteredAlerts,
		"count":   len(filteredAlerts),
		"user_id": userID,
	})
}

// MarkAlertAsRead handles PUT /api/alerts/:id/read - mark an alert as read
func (h *AlertHandler) MarkAlertAsRead(c *gin.Context) {
	idParam := c.Param("id")
	alertID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert ID",
			"details": "Alert ID must be a valid integer",
		})
		return
	}

	if err := h.alertService.MarkAlertAsRead(alertID); err != nil {
		if err.Error() == "alert not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Alert not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark alert as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert marked as read successfully",
	})
}

// MarkAllUserAlertsAsRead handles PUT /api/alerts/user/:id/read-all - mark all user alerts as read
func (h *AlertHandler) MarkAllUserAlertsAsRead(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	if err := h.alertService.MarkAllUserAlertsAsRead(userID); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark alerts as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All user alerts marked as read successfully",
	})
}

// DeleteAlert handles DELETE /api/alerts/:id - delete an alert
func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	idParam := c.Param("id")
	alertID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert ID",
			"details": "Alert ID must be a valid integer",
		})
		return
	}

	if err := h.alertService.DeleteAlert(alertID); err != nil {
		if err.Error() == "alert not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Alert not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert deleted successfully",
	})
}

// GetAlertsSummary handles GET /api/alerts/summary - get alerts summary grouped by user
func (h *AlertHandler) GetAlertsSummary(c *gin.Context) {
	summary, err := h.alertService.GetAlertsSummaryByUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve alerts summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
		"users":   len(summary),
	})
}

// GetDashboard handles GET /api/alerts/dashboard - get dashboard with overdue items and upcoming due dates
func (h *AlertHandler) GetDashboard(c *gin.Context) {
	// Get alerts summary
	summary, err := h.alertService.GetAlertsSummaryByUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve dashboard data",
			"details": err.Error(),
		})
		return
	}

	// Calculate totals
	totalAlerts := 0
	totalOverdue := 0
	totalReminders := 0
	usersWithAlerts := len(summary)

	for _, userSummary := range summary {
		totalAlerts += userSummary.TotalAlerts
		totalOverdue += userSummary.OverdueCount
		totalReminders += userSummary.ReminderCount
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboard": gin.H{
			"total_alerts":      totalAlerts,
			"total_overdue":     totalOverdue,
			"total_reminders":   totalReminders,
			"users_with_alerts": usersWithAlerts,
		},
		"user_summaries": summary,
	})
}

// CreateCustomAlertRequest represents the request body for creating a custom alert
type CreateCustomAlertRequest struct {
	UserID    int    `json:"user_id" binding:"required"`
	GameID    int    `json:"game_id" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

// CreateCustomAlert handles POST /api/alerts - create a custom alert
func (h *AlertHandler) CreateCustomAlert(c *gin.Context) {
	var req CreateCustomAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	alert, err := h.alertService.CreateCustomAlert(req.UserID, req.GameID, req.Type, req.Message)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Resource not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Alert created successfully",
		"alert":   alert,
	})
}

// GenerateOverdueAlerts handles POST /api/alerts/generate-overdue - generate overdue alerts
func (h *AlertHandler) GenerateOverdueAlerts(c *gin.Context) {
	if err := h.alertService.GenerateOverdueAlerts(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate overdue alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Overdue alerts generated successfully",
	})
}

// GenerateReminderAlerts handles POST /api/alerts/generate-reminders - generate reminder alerts
func (h *AlertHandler) GenerateReminderAlerts(c *gin.Context) {
	if err := h.alertService.GenerateReminderAlerts(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate reminder alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reminder alerts generated successfully",
	})
}

// CleanupResolvedAlerts handles POST /api/alerts/cleanup - cleanup resolved alerts
func (h *AlertHandler) CleanupResolvedAlerts(c *gin.Context) {
	if err := h.alertService.CleanupResolvedAlerts(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup resolved alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Resolved alerts cleaned up successfully",
	})
}

// RegisterRoutes registers all alert-related routes
func (h *AlertHandler) RegisterRoutes(router *gin.RouterGroup) {
	alerts := router.Group("/alerts")
	{
		alerts.GET("", h.GetActiveAlerts)
		alerts.POST("", h.CreateCustomAlert)
		alerts.GET("/summary", h.GetAlertsSummary)
		alerts.GET("/dashboard", h.GetDashboard)
		alerts.GET("/user/:id", h.GetAlertsByUser)
		alerts.PUT("/user/:id/read-all", h.MarkAllUserAlertsAsRead)
		alerts.PUT("/:id/read", h.MarkAlertAsRead)
		alerts.DELETE("/:id", h.DeleteAlert)
		alerts.POST("/generate-overdue", h.GenerateOverdueAlerts)
		alerts.POST("/generate-reminders", h.GenerateReminderAlerts)
		alerts.POST("/cleanup", h.CleanupResolvedAlerts)
	}
}