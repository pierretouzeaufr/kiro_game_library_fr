package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDashboardWebHandler_Creation(t *testing.T) {
	mockAlertService := new(MockAlertService)
	mockBorrowingService := new(MockBorrowingService)
	mockGameService := new(MockGameService)
	mockUserService := new(MockUserService)

	handler := NewDashboardWebHandler(mockAlertService, mockBorrowingService, mockGameService, mockUserService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.alertService)
	assert.NotNil(t, handler.borrowingService)
	assert.NotNil(t, handler.gameService)
	assert.NotNil(t, handler.userService)
}

func TestDashboardWebHandler_RegisterWebRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAlertService := new(MockAlertService)
	mockBorrowingService := new(MockBorrowingService)
	mockGameService := new(MockGameService)
	mockUserService := new(MockUserService)

	handler := NewDashboardWebHandler(mockAlertService, mockBorrowingService, mockGameService, mockUserService)
	
	router := gin.New()
	routerGroup := router.Group("/")
	
	// This should not panic
	assert.NotPanics(t, func() {
		handler.RegisterWebRoutes(routerGroup)
	})
	
	// Check that routes were registered
	routes := router.Routes()
	assert.True(t, len(routes) > 0, "Routes should be registered")
	
	// Check for specific routes
	routePaths := make([]string, len(routes))
	for i, route := range routes {
		routePaths[i] = route.Path
	}
	
	expectedRoutes := []string{
		"/dashboard",
		"/dashboard/stats",
		"/dashboard/alerts",
		"/dashboard/content",
	}
	
	for _, expectedRoute := range expectedRoutes {
		found := false
		for _, actualRoute := range routePaths {
			if actualRoute == expectedRoute {
				found = true
				break
			}
		}
		assert.True(t, found, "Route %s should be registered", expectedRoute)
	}
}