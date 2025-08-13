package handlers

import (
	"board-game-library/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// BorrowingServiceInterface defines the interface for borrowing service operations
type BorrowingServiceInterface interface {
	BorrowGame(userID, gameID int, dueDate time.Time) (*models.Borrowing, error)
	BorrowGameWithDefaultDueDate(userID, gameID int) (*models.Borrowing, error)
	ReturnGame(borrowingID int) error
	GetOverdueItems() ([]*models.Borrowing, error)
	ExtendDueDate(borrowingID int, newDueDate time.Time) error
	GetBorrowingDetails(borrowingID int) (*models.Borrowing, error)
	GetActiveBorrowingsByUser(userID int) ([]*models.Borrowing, error)
	GetBorrowingsByGame(gameID int) ([]*models.Borrowing, error)
	UpdateOverdueStatus() error
	GetItemsDueSoon(daysAhead int) ([]*models.Borrowing, error)
}

// BorrowingHandler handles HTTP requests for borrowing workflow
type BorrowingHandler struct {
	borrowingService BorrowingServiceInterface
}

// NewBorrowingHandler creates a new BorrowingHandler instance
func NewBorrowingHandler(borrowingService BorrowingServiceInterface) *BorrowingHandler {
	return &BorrowingHandler{
		borrowingService: borrowingService,
	}
}

// BorrowGameRequest represents the request body for borrowing a game
type BorrowGameRequest struct {
	UserID  int    `json:"user_id" binding:"required"`
	GameID  int    `json:"game_id" binding:"required"`
	DueDate string `json:"due_date,omitempty"` // Optional, format: "2006-01-02"
}

// BorrowGame handles POST /api/borrowings - borrow a game
func (h *BorrowingHandler) BorrowGame(c *gin.Context) {
	var req BorrowGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	var borrowing *models.Borrowing
	var err error

	if req.DueDate == "" {
		// Use default due date (14 days)
		borrowing, err = h.borrowingService.BorrowGameWithDefaultDueDate(req.UserID, req.GameID)
	} else {
		// Parse custom due date
		dueDate, parseErr := time.Parse("2006-01-02", req.DueDate)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid due date format",
				"details": "Due date must be in format YYYY-MM-DD",
			})
			return
		}
		borrowing, err = h.borrowingService.BorrowGame(req.UserID, req.GameID, dueDate)
	}

	if err != nil {
		if err.Error() == "user not found" || err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Resource not found",
				"details": err.Error(),
			})
			return
		}
		if err.Error() == "game is not available for borrowing" || 
		   err.Error() == "user has overdue items and cannot borrow" ||
		   err.Error() == "user account is inactive" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Cannot borrow game",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to borrow game",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Game borrowed successfully",
		"borrowing": borrowing,
	})
}

// ReturnGame handles PUT /api/borrowings/:id/return - return a borrowed game
func (h *BorrowingHandler) ReturnGame(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid borrowing ID",
			"details": "Borrowing ID must be a valid integer",
		})
		return
	}

	if err := h.borrowingService.ReturnGame(id); err != nil {
		if err.Error() == "borrowing not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Borrowing not found",
				"details": err.Error(),
			})
			return
		}
		if err.Error() == "game has already been returned" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Game already returned",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to return game",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game returned successfully",
	})
}

// GetBorrowingDetails handles GET /api/borrowings/:id - get borrowing details
func (h *BorrowingHandler) GetBorrowingDetails(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid borrowing ID",
			"details": "Borrowing ID must be a valid integer",
		})
		return
	}

	borrowing, err := h.borrowingService.GetBorrowingDetails(id)
	if err != nil {
		if err.Error() == "borrowing not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Borrowing not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve borrowing details",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"borrowing": borrowing,
	})
}

// ExtendDueDateRequest represents the request body for extending due date
type ExtendDueDateRequest struct {
	NewDueDate string `json:"new_due_date" binding:"required"` // Format: "2006-01-02"
}

// ExtendDueDate handles PUT /api/borrowings/:id/extend - extend due date
func (h *BorrowingHandler) ExtendDueDate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid borrowing ID",
			"details": "Borrowing ID must be a valid integer",
		})
		return
	}

	var req ExtendDueDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Parse new due date
	newDueDate, err := time.Parse("2006-01-02", req.NewDueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid due date format",
			"details": "Due date must be in format YYYY-MM-DD",
		})
		return
	}

	if err := h.borrowingService.ExtendDueDate(id, newDueDate); err != nil {
		if err.Error() == "borrowing not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Borrowing not found",
				"details": err.Error(),
			})
			return
		}
		if err.Error() == "cannot extend due date for returned item" ||
		   err.Error() == "new due date must be after borrowed date" ||
		   err.Error() == "due date cannot be more than 90 days from borrowed date" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid due date extension",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to extend due date",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Due date extended successfully",
	})
}

// GetOverdueItems handles GET /api/borrowings/overdue - get all overdue items
func (h *BorrowingHandler) GetOverdueItems(c *gin.Context) {
	overdueItems, err := h.borrowingService.GetOverdueItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve overdue items",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"overdue_items": overdueItems,
		"count":         len(overdueItems),
	})
}

// GetItemsDueSoon handles GET /api/borrowings/due-soon - get items due soon
func (h *BorrowingHandler) GetItemsDueSoon(c *gin.Context) {
	daysParam := c.DefaultQuery("days", "2") // Default to 2 days
	days, err := strconv.Atoi(daysParam)
	if err != nil || days < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid days parameter",
			"details": "Days must be a non-negative integer",
		})
		return
	}

	itemsDueSoon, err := h.borrowingService.GetItemsDueSoon(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve items due soon",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items_due_soon": itemsDueSoon,
		"count":          len(itemsDueSoon),
		"days_ahead":     days,
	})
}

// GetActiveBorrowingsByUser handles GET /api/borrowings/user/:id - get active borrowings by user
func (h *BorrowingHandler) GetActiveBorrowingsByUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
		return
	}

	borrowings, err := h.borrowingService.GetActiveBorrowingsByUser(userID)
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
		"user_id":    userID,
	})
}

// GetBorrowingsByGame handles GET /api/borrowings/game/:id - get borrowings by game
func (h *BorrowingHandler) GetBorrowingsByGame(c *gin.Context) {
	idParam := c.Param("id")
	gameID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid game ID",
			"details": "Game ID must be a valid integer",
		})
		return
	}

	borrowings, err := h.borrowingService.GetBorrowingsByGame(gameID)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Game not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve game borrowings",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"borrowings": borrowings,
		"count":      len(borrowings),
		"game_id":    gameID,
	})
}

// UpdateOverdueStatus handles POST /api/borrowings/update-overdue - update overdue status
func (h *BorrowingHandler) UpdateOverdueStatus(c *gin.Context) {
	if err := h.borrowingService.UpdateOverdueStatus(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update overdue status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Overdue status updated successfully",
	})
}

// RegisterRoutes registers all borrowing-related routes
func (h *BorrowingHandler) RegisterRoutes(router *gin.RouterGroup) {
	borrowings := router.Group("/borrowings")
	{
		borrowings.POST("", h.BorrowGame)
		borrowings.GET("/:id", h.GetBorrowingDetails)
		borrowings.PUT("/:id/return", h.ReturnGame)
		borrowings.PUT("/:id/extend", h.ExtendDueDate)
		borrowings.GET("/overdue", h.GetOverdueItems)
		borrowings.GET("/due-soon", h.GetItemsDueSoon)
		borrowings.GET("/user/:id", h.GetActiveBorrowingsByUser)
		borrowings.GET("/game/:id", h.GetBorrowingsByGame)
		borrowings.POST("/update-overdue", h.UpdateOverdueStatus)
	}
}