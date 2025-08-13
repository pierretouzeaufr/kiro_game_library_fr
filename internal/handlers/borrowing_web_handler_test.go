package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBorrowingWebHandler_Creation(t *testing.T) {
	mockBorrowingService := new(MockBorrowingService)
	mockUserService := new(MockUserService)
	mockGameService := new(MockGameService)

	handler := NewBorrowingWebHandler(mockBorrowingService, mockUserService, mockGameService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.borrowingService)
	assert.NotNil(t, handler.userService)
	assert.NotNil(t, handler.gameService)
}

func TestBorrowingWebHandler_RegisterWebRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockBorrowingService := new(MockBorrowingService)
	mockUserService := new(MockUserService)
	mockGameService := new(MockGameService)

	handler := NewBorrowingWebHandler(mockBorrowingService, mockUserService, mockGameService)
	
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
		"/borrowings/new",
		"/borrowings",
		"/borrowings/:id",
		"/borrowings/:id/return",
		"/borrowings/:id/extend",
		"/games/:id/availability",
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