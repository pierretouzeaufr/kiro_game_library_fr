package handlers

import (
	"board-game-library/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserServiceInterface defines the interface for user service operations
type UserServiceInterface interface {
	RegisterUser(name, email string) (*models.User, error)
	GetUser(id int) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	GetUserBorrowings(userID int) ([]*models.Borrowing, error)
	CanUserBorrow(userID int) (bool, error)
	GetActiveUserBorrowings(userID int) ([]*models.Borrowing, error)
	UpdateUser(user *models.User) error
}

// UserHandler handles HTTP requests for user management
type UserHandler struct {
	userService UserServiceInterface
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterUserRequest represents the request body for user registration
type RegisterUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// RegisterUser handles POST /api/users - user registration
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.RegisterUser(req.Name, req.Email)
	if err != nil {
		if err.Error() == "user with email "+req.Email+" already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "User already exists",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to register user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// GetAllUsers handles GET /api/users - list all users
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// GetUser handles GET /api/users/:id - get user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// GetUserBorrowings handles GET /api/users/:id/borrowings - get user borrowing history
func (h *UserHandler) GetUserBorrowings(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	borrowings, err := h.userService.GetUserBorrowings(id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user borrowings",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"borrowings": borrowings,
		"count":      len(borrowings),
	})
}

// GetUserCurrentLoans handles GET /api/users/:id/current-loans - get user's current active borrowings
func (h *UserHandler) GetUserCurrentLoans(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	activeBorrowings, err := h.userService.GetActiveUserBorrowings(id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve current loans",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current_loans": activeBorrowings,
		"count":         len(activeBorrowings),
	})
}

// UpdateUserRequest represents the request body for user updates
type UpdateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	IsActive *bool  `json:"is_active"`
}

// UpdateUser handles PUT /api/users/:id - update user information
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get existing user
	existingUser, err := h.userService.GetUser(id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user",
			"details": err.Error(),
		})
		return
	}

	// Update user fields
	existingUser.Name = req.Name
	existingUser.Email = req.Email
	if req.IsActive != nil {
		existingUser.IsActive = *req.IsActive
	}

	// Update user
	if err := h.userService.UpdateUser(existingUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    existingUser,
	})
}

// CheckUserEligibility handles GET /api/users/:id/eligibility - check if user can borrow
func (h *UserHandler) CheckUserEligibility(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	canBorrow, err := h.userService.CanUserBorrow(id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"details": err.Error(),
			})
			return
		}
		// For business logic errors (like overdue items), return success with can_borrow=false
		c.JSON(http.StatusOK, gin.H{
			"can_borrow": false,
			"reason":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"can_borrow": canBorrow,
		"reason":     "User is eligible to borrow",
	})
}

// RegisterRoutes registers all user-related routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.POST("", h.RegisterUser)
		users.GET("", h.GetAllUsers)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.GET("/:id/borrowings", h.GetUserBorrowings)
		users.GET("/:id/current-loans", h.GetUserCurrentLoans)
		users.GET("/:id/eligibility", h.CheckUserEligibility)
	}
}