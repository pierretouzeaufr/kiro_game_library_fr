package handlers

import (
	"board-game-library/internal/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AlertWebHandler handles web requests for alert management with HTMX support
type AlertWebHandler struct {
	alertService AlertServiceInterface
}

// NewAlertWebHandler creates a new AlertWebHandler instance
func NewAlertWebHandler(alertService AlertServiceInterface) *AlertWebHandler {
	return &AlertWebHandler{
		alertService: alertService,
	}
}

// ShowAlertsList handles GET /alerts - display alerts list page
func (h *AlertWebHandler) ShowAlertsList(c *gin.Context) {
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load alerts: " + err.Error(),
		})
		return
	}

	// Group alerts by user
	groupedAlerts := make(map[int][]*models.Alert)
	stats := struct {
		OverdueAlerts   int
		ReminderAlerts  int
		UnreadAlerts    int
	}{}

	for _, alert := range alerts {
		groupedAlerts[alert.UserID] = append(groupedAlerts[alert.UserID], alert)
		
		if alert.Type == "overdue" {
			stats.OverdueAlerts++
		} else if alert.Type == "reminder" {
			stats.ReminderAlerts++
		}
		
		if !alert.IsRead {
			stats.UnreadAlerts++
		}
	}

	c.HTML(http.StatusOK, "alerts/list.html", gin.H{
		"GroupedAlerts": groupedAlerts,
		"Stats":         stats,
		"HasFilters":    false,
	})
}

// MarkAlertAsRead handles POST /alerts/:id/mark-read - mark alert as read with HTMX response
func (h *AlertWebHandler) MarkAlertAsRead(c *gin.Context) {
	idParam := c.Param("id")
	alertID, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid alert ID",
		})
		return
	}

	if err := h.alertService.MarkAlertAsRead(alertID); err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to mark alert as read: " + err.Error(),
		})
		return
	}

	// Return updated alerts list
	c.Header("HX-Trigger", "alert-marked-read, refresh-dashboard")
	h.renderAlertsList(c)
}

// MarkAllAlertsAsRead handles POST /alerts/mark-all-read - mark all alerts as read
func (h *AlertWebHandler) MarkAllAlertsAsRead(c *gin.Context) {
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load alerts: " + err.Error(),
		})
		return
	}

	// Mark all unread alerts as read
	for _, alert := range alerts {
		if !alert.IsRead {
			h.alertService.MarkAlertAsRead(alert.ID)
		}
	}

	// Return updated alerts list
	c.Header("HX-Trigger", "all-alerts-marked-read, refresh-dashboard")
	h.renderAlertsList(c)
}

// MarkUserAlertsAsRead handles POST /alerts/mark-user-read/:id - mark all user alerts as read
func (h *AlertWebHandler) MarkUserAlertsAsRead(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid user ID",
		})
		return
	}

	if err := h.alertService.MarkAllUserAlertsAsRead(userID); err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to mark user alerts as read: " + err.Error(),
		})
		return
	}

	// Return updated alerts list
	c.Header("HX-Trigger", "user-alerts-marked-read, refresh-dashboard")
	h.renderAlertsList(c)
}

// GenerateAlerts handles POST /alerts/generate - generate new alerts
func (h *AlertWebHandler) GenerateAlerts(c *gin.Context) {
	// Generate overdue alerts
	if err := h.alertService.GenerateOverdueAlerts(); err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to generate overdue alerts: " + err.Error(),
		})
		return
	}

	// Generate reminder alerts
	if err := h.alertService.GenerateReminderAlerts(); err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to generate reminder alerts: " + err.Error(),
		})
		return
	}

	// Return updated alerts list
	c.Header("HX-Trigger", "alerts-generated, refresh-dashboard")
	h.renderAlertsList(c)
}

// SearchAlerts handles POST /alerts/search - search alerts with HTMX
func (h *AlertWebHandler) SearchAlerts(c *gin.Context) {
	query := strings.TrimSpace(c.PostForm("search"))
	
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Search failed: " + err.Error(),
		})
		return
	}

	// Filter alerts based on search query
	var filteredAlerts []*models.Alert
	if query == "" {
		filteredAlerts = alerts
	} else {
		queryLower := strings.ToLower(query)
		for _, alert := range alerts {
			if strings.Contains(strings.ToLower(alert.Message), queryLower) {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
	}

	h.renderFilteredAlertsList(c, filteredAlerts, true)
}

// FilterAlerts handles POST /alerts/filter - filter alerts with HTMX
func (h *AlertWebHandler) FilterAlerts(c *gin.Context) {
	alertType := c.PostForm("type")
	status := c.PostForm("status")
	
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Filter failed: " + err.Error(),
		})
		return
	}

	// Apply filters
	var filteredAlerts []*models.Alert
	for _, alert := range alerts {
		// Filter by type
		if alertType != "" && alertType != "all" && alert.Type != alertType {
			continue
		}
		
		// Filter by status
		if status != "" && status != "all" {
			if status == "unread" && alert.IsRead {
				continue
			}
			if status == "read" && !alert.IsRead {
				continue
			}
		}
		
		filteredAlerts = append(filteredAlerts, alert)
	}

	hasFilters := (alertType != "" && alertType != "all") || (status != "" && status != "all")
	h.renderFilteredAlertsList(c, filteredAlerts, hasFilters)
}

// GetAlertCount handles GET /alerts/count - get alert count for dynamic updates
func (h *AlertWebHandler) GetAlertCount(c *gin.Context) {
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to get alert count: " + err.Error(),
		})
		return
	}

	unreadCount := 0
	for _, alert := range alerts {
		if !alert.IsRead {
			unreadCount++
		}
	}

	c.HTML(http.StatusOK, "alerts/partials/alert-count.html", gin.H{
		"UnreadCount": unreadCount,
		"TotalCount":  len(alerts),
	})
}

// renderAlertsList renders the complete alerts list
func (h *AlertWebHandler) renderAlertsList(c *gin.Context) {
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load alerts: " + err.Error(),
		})
		return
	}

	h.renderFilteredAlertsList(c, alerts, false)
}

// renderFilteredAlertsList renders the filtered alerts list
func (h *AlertWebHandler) renderFilteredAlertsList(c *gin.Context, alerts []*models.Alert, hasFilters bool) {
	// Group alerts by user
	groupedAlerts := make(map[int][]*models.Alert)
	stats := struct {
		OverdueAlerts   int
		ReminderAlerts  int
		UnreadAlerts    int
	}{}

	for _, alert := range alerts {
		groupedAlerts[alert.UserID] = append(groupedAlerts[alert.UserID], alert)
		
		if alert.Type == "overdue" {
			stats.OverdueAlerts++
		} else if alert.Type == "reminder" {
			stats.ReminderAlerts++
		}
		
		if !alert.IsRead {
			stats.UnreadAlerts++
		}
	}

	c.HTML(http.StatusOK, "alerts/partials/alerts-list.html", gin.H{
		"GroupedAlerts": groupedAlerts,
		"Stats":         stats,
		"HasFilters":    hasFilters,
	})
}

// RegisterWebRoutes registers all alert web routes
func (h *AlertWebHandler) RegisterWebRoutes(router *gin.RouterGroup) {
	alerts := router.Group("/alerts")
	{
		alerts.GET("", h.ShowAlertsList)
		alerts.POST("/search", h.SearchAlerts)
		alerts.POST("/filter", h.FilterAlerts)
		alerts.POST("/mark-all-read", h.MarkAllAlertsAsRead)
		alerts.POST("/mark-user-read/:id", h.MarkUserAlertsAsRead)
		alerts.POST("/:id/mark-read", h.MarkAlertAsRead)
		alerts.POST("/generate", h.GenerateAlerts)
		alerts.GET("/count", h.GetAlertCount)
	}
}