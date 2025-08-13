package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DashboardWebHandler handles web requests for dashboard with HTMX support
type DashboardWebHandler struct {
	alertService     AlertServiceInterface
	borrowingService BorrowingServiceInterface
	gameService      GameServiceInterface
	userService      UserServiceInterface
}

// NewDashboardWebHandler creates a new DashboardWebHandler instance
func NewDashboardWebHandler(alertService AlertServiceInterface, borrowingService BorrowingServiceInterface, gameService GameServiceInterface, userService UserServiceInterface) *DashboardWebHandler {
	return &DashboardWebHandler{
		alertService:     alertService,
		borrowingService: borrowingService,
		gameService:      gameService,
		userService:      userService,
	}
}

// ShowDashboard handles GET /dashboard - display dashboard page
func (h *DashboardWebHandler) ShowDashboard(c *gin.Context) {
	dashboardData, err := h.getDashboardData()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load dashboard: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", dashboardData)
}

// GetDashboardStats handles GET /dashboard/stats - get dashboard stats for HTMX updates
func (h *DashboardWebHandler) GetDashboardStats(c *gin.Context) {
	dashboardData, err := h.getDashboardData()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load dashboard stats: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard/partials/stats.html", dashboardData)
}

// GetDashboardAlerts handles GET /dashboard/alerts - get dashboard alerts section for HTMX updates
func (h *DashboardWebHandler) GetDashboardAlerts(c *gin.Context) {
	dashboardData, err := h.getDashboardData()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load dashboard alerts: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard/partials/alerts.html", dashboardData)
}

// GetDashboardContent handles GET /dashboard/content - get dashboard main content for HTMX updates
func (h *DashboardWebHandler) GetDashboardContent(c *gin.Context) {
	dashboardData, err := h.getDashboardData()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load dashboard content: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard/partials/content.html", dashboardData)
}

// getDashboardData collects all dashboard data
func (h *DashboardWebHandler) getDashboardData() (gin.H, error) {
	// Get basic stats
	games, err := h.gameService.GetAllGames()
	if err != nil {
		return nil, err
	}

	users, err := h.userService.GetAllUsers()
	if err != nil {
		return nil, err
	}

	// Get overdue items
	overdueItems, err := h.borrowingService.GetOverdueItems()
	if err != nil {
		return nil, err
	}

	// Get items due soon (within 2 days)
	dueSoonItems, err := h.borrowingService.GetItemsDueSoon(2)
	if err != nil {
		return nil, err
	}

	// Count active borrowings
	activeBorrowings := 0
	for _, game := range games {
		if !game.IsAvailable {
			activeBorrowings++
		}
	}

	// Count active users (users with current borrowings)
	activeUsers := 0
	for _, user := range users {
		if userBorrowings, err := h.borrowingService.GetActiveBorrowingsByUser(user.ID); err == nil && len(userBorrowings) > 0 {
			activeUsers++
		}
	}

	// Create recent activity (simplified)
	recentActivity := []gin.H{}
	
	// Add recent borrowings to activity
	for i, borrowing := range overdueItems {
		if i >= 5 { // Limit to 5 items
			break
		}
		recentActivity = append(recentActivity, gin.H{
			"Description":  "Overdue item needs attention",
			"Timestamp":    borrowing.DueDate,
			"IconBgColor":  "bg-red-500",
			"IconColor":    "text-white",
			"IconPath":     `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>`,
			"IsLast":       i == len(overdueItems)-1 && len(dueSoonItems) == 0,
		})
	}

	// Add due soon items to activity
	for i, borrowing := range dueSoonItems {
		if len(recentActivity) >= 5 { // Limit total to 5 items
			break
		}
		recentActivity = append(recentActivity, gin.H{
			"Description":  "Item due soon",
			"Timestamp":    borrowing.DueDate,
			"IconBgColor":  "bg-yellow-500",
			"IconColor":    "text-white",
			"IconPath":     `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>`,
			"IsLast":       i == len(dueSoonItems)-1,
		})
	}

	return gin.H{
		"Stats": gin.H{
			"TotalGames":       len(games),
			"ActiveUsers":      activeUsers,
			"ActiveBorrowings": activeBorrowings,
			"OverdueItems":     len(overdueItems),
		},
		"OverdueItems":    overdueItems,
		"DueSoonItems":    dueSoonItems,
		"RecentActivity":  recentActivity,
	}, nil
}

// RegisterWebRoutes registers all dashboard web routes
func (h *DashboardWebHandler) RegisterWebRoutes(router *gin.RouterGroup) {
	dashboard := router.Group("/dashboard")
	{
		dashboard.GET("", h.ShowDashboard)
		dashboard.GET("/stats", h.GetDashboardStats)
		dashboard.GET("/alerts", h.GetDashboardAlerts)
		dashboard.GET("/content", h.GetDashboardContent)
	}
}