package handlers

import (
	"board-game-library/internal/models"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserWebHandler handles web requests for user management with HTMX support
type UserWebHandler struct {
	userService UserServiceInterface
}

// NewUserWebHandler creates a new UserWebHandler instance
func NewUserWebHandler(userService UserServiceInterface) *UserWebHandler {
	return &UserWebHandler{
		userService: userService,
	}
}

// ListUsers handles GET /users - display users list page
func (h *UserWebHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Title":        "Error",
			"ErrorMessage": "Failed to load users: " + err.Error(),
		})
		return
	}

	// Add current loans count for each user
	for _, user := range users {
		activeBorrowings, _ := h.userService.GetActiveUserBorrowings(user.ID)
		user.CurrentLoans = len(activeBorrowings)
	}

	c.HTML(http.StatusOK, "users/list.html", gin.H{
		"Title": "Users",
		"Users": users,
	})
}

// SearchUsers handles POST /users/search - HTMX search for users
func (h *UserWebHandler) SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.PostForm("search"))
	
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Search failed: " + err.Error(),
		})
		return
	}

	// Filter users based on search query
	if query != "" {
		filteredUsers := make([]*models.User, 0)
		queryLower := strings.ToLower(query)
		
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.Name), queryLower) ||
			   strings.Contains(strings.ToLower(user.Email), queryLower) {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
	}

	// Add current loans count for each user
	for _, user := range users {
		activeBorrowings, _ := h.userService.GetActiveUserBorrowings(user.ID)
		user.CurrentLoans = len(activeBorrowings)
	}

	c.HTML(http.StatusOK, "users/partials/users-table.html", gin.H{
		"Users": users,
	})
}

// SearchFilterUsers handles POST /users/search-filter - Combined HTMX search and filter
func (h *UserWebHandler) SearchFilterUsers(c *gin.Context) {
	query := strings.TrimSpace(c.PostForm("search"))
	status := c.PostForm("status")
	
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Search failed: " + err.Error(),
		})
		return
	}

	// Add current loans count for each user first
	for _, user := range users {
		activeBorrowings, _ := h.userService.GetActiveUserBorrowings(user.ID)
		user.CurrentLoans = len(activeBorrowings)
	}

	// Filter users based on search query
	if query != "" {
		filteredUsers := make([]*models.User, 0)
		queryLower := strings.ToLower(query)
		
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.Name), queryLower) ||
			   strings.Contains(strings.ToLower(user.Email), queryLower) {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
	}

	// Filter by status
	if status != "" && status != "all" {
		filteredUsers := make([]*models.User, 0)
		for _, user := range users {
			switch status {
			case "active":
				if user.IsActive {
					filteredUsers = append(filteredUsers, user)
				}
			case "inactive":
				if !user.IsActive {
					filteredUsers = append(filteredUsers, user)
				}
			case "with-loans":
				if user.CurrentLoans > 0 {
					filteredUsers = append(filteredUsers, user)
				}
			}
		}
		users = filteredUsers
	}

	c.HTML(http.StatusOK, "users/partials/users-table.html", gin.H{
		"Users": users,
	})
}

// SortUsers handles POST /users/sort - HTMX sort for users
func (h *UserWebHandler) SortUsers(c *gin.Context) {
	field := c.PostForm("field")
	direction := c.PostForm("direction")
	
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Sort failed: " + err.Error(),
		})
		return
	}

	// Add current loans count for each user
	for _, user := range users {
		activeBorrowings, _ := h.userService.GetActiveUserBorrowings(user.ID)
		user.CurrentLoans = len(activeBorrowings)
	}

	// Sort users based on field and direction
	sortUsers(users, field, direction)

	c.HTML(http.StatusOK, "users/partials/users-table.html", gin.H{
		"Users":         users,
		"SortField":     field,
		"SortDirection": direction,
	})
}

// ShowUser handles GET /users/:id - display user details modal
func (h *UserWebHandler) ShowUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "User not found",
		})
		return
	}

	// Get borrowing history
	borrowings, _ := h.userService.GetUserBorrowings(id)
	activeBorrowings, _ := h.userService.GetActiveUserBorrowings(id)

	c.HTML(http.StatusOK, "users/detail.html", gin.H{
		"User":             user,
		"Borrowings":       borrowings,
		"ActiveBorrowings": activeBorrowings,
	})
}

// ShowNewUserForm handles GET /users/new - display new user form modal
func (h *UserWebHandler) ShowNewUserForm(c *gin.Context) {
	c.HTML(http.StatusOK, "users/new.html", gin.H{
		"Title": "Add New User",
	})
}

// ShowEditUserForm handles GET /users/:id/edit - display edit user form modal
func (h *UserWebHandler) ShowEditUserForm(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "User not found",
		})
		return
	}

	c.HTML(http.StatusOK, "users/edit.html", gin.H{
		"User": user,
	})
}

// RegisterWebRoutes registers all user web routes
func (h *UserWebHandler) RegisterWebRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.GET("", h.ListUsers)
		users.POST("/search", h.SearchUsers)
		users.POST("/search-filter", h.SearchFilterUsers)
		users.POST("/sort", h.SortUsers)
		users.GET("/new", h.ShowNewUserForm)
		users.GET("/:id", h.ShowUser)
		users.GET("/:id/edit", h.ShowEditUserForm)
	}
}

// sortUsers sorts users slice based on field and direction
func sortUsers(users []*models.User, field, direction string) {
	switch field {
	case "name":
		sort.Slice(users, func(i, j int) bool {
			if direction == "desc" {
				return users[i].Name > users[j].Name
			}
			return users[i].Name < users[j].Name
		})
	case "registered_at":
		sort.Slice(users, func(i, j int) bool {
			if direction == "desc" {
				return users[i].RegisteredAt.After(users[j].RegisteredAt)
			}
			return users[i].RegisteredAt.Before(users[j].RegisteredAt)
		})
	case "email":
		sort.Slice(users, func(i, j int) bool {
			if direction == "desc" {
				return users[i].Email > users[j].Email
			}
			return users[i].Email < users[j].Email
		})
	}
}