package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAlertWebHandler_Creation(t *testing.T) {
	mockAlertService := new(MockAlertService)
	handler := NewAlertWebHandler(mockAlertService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.alertService)
}

func TestAlertWebHandler_RegisterWebRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAlertService := new(MockAlertService)
	handler := NewAlertWebHandler(mockAlertService)
	
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
		"/alerts",
		"/alerts/search",
		"/alerts/filter",
		"/alerts/mark-all-read",
		"/alerts/mark-user-read/:id",
		"/alerts/:id/mark-read",
		"/alerts/generate",
		"/alerts/count",
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