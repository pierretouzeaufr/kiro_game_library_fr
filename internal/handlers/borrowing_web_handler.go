package handlers

import (
	"board-game-library/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// BorrowingWebHandler handles web requests for borrowing workflow with HTMX support
type BorrowingWebHandler struct {
	borrowingService BorrowingServiceInterface
	userService      UserServiceInterface
	gameService      GameServiceInterface
}

// NewBorrowingWebHandler creates a new BorrowingWebHandler instance
func NewBorrowingWebHandler(borrowingService BorrowingServiceInterface, userService UserServiceInterface, gameService GameServiceInterface) *BorrowingWebHandler {
	return &BorrowingWebHandler{
		borrowingService: borrowingService,
		userService:      userService,
		gameService:      gameService,
	}
}

// ShowNewBorrowingForm handles GET /borrowings/new - display new borrowing form modal
func (h *BorrowingWebHandler) ShowNewBorrowingForm(c *gin.Context) {
	gameIDParam := c.Query("game_id")
	
	// Get available games
	availableGames, err := h.gameService.GetAvailableGames()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load available games: " + err.Error(),
		})
		return
	}

	// Get eligible users (active users without overdue items)
	allUsers, err := h.userService.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to load users: " + err.Error(),
		})
		return
	}

	eligibleUsers := make([]*models.User, 0)
	for _, user := range allUsers {
		if canBorrow, _ := h.userService.CanUserBorrow(user.ID); canBorrow {
			eligibleUsers = append(eligibleUsers, user)
		}
	}

	templateData := gin.H{
		"AvailableGames": availableGames,
		"EligibleUsers":  eligibleUsers,
		"DefaultDueDate": time.Now().AddDate(0, 0, 14).Format("2006-01-02"),
		"MinDate":        time.Now().Format("2006-01-02"),
		"MaxDate":        time.Now().AddDate(0, 0, 90).Format("2006-01-02"),
	}

	// If game_id is provided, pre-select the game
	if gameIDParam != "" {
		if gameID, err := strconv.Atoi(gameIDParam); err == nil {
			if game, err := h.gameService.GetGame(gameID); err == nil && game.IsAvailable {
				templateData["Game"] = game
			}
		}
	}

	c.HTML(http.StatusOK, "borrowings/new.html", templateData)
}

// CreateBorrowing handles POST /borrowings - create new borrowing with HTMX response
func (h *BorrowingWebHandler) CreateBorrowing(c *gin.Context) {
	userIDStr := c.PostForm("user_id")
	gameIDStr := c.PostForm("game_id")
	dueDateStr := c.PostForm("due_date")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid user ID",
		})
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid game ID",
		})
		return
	}

	var borrowing *models.Borrowing
	if dueDateStr == "" {
		// Use default due date
		borrowing, err = h.borrowingService.BorrowGameWithDefaultDueDate(userID, gameID)
	} else {
		dueDate, parseErr := time.Parse("2006-01-02", dueDateStr)
		if parseErr != nil {
			c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
				"ErrorMessage": "Invalid due date format",
			})
			return
		}
		borrowing, err = h.borrowingService.BorrowGame(userID, gameID, dueDate)
	}

	if err != nil {
		c.HTML(http.StatusConflict, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to borrow game: " + err.Error(),
		})
		return
	}

	// Return success message with updated game status
	c.Header("HX-Trigger", "borrowing-created, refresh-games, refresh-dashboard")
	c.HTML(http.StatusOK, "borrowings/success.html", gin.H{
		"Message":   "Game borrowed successfully!",
		"Borrowing": borrowing,
	})
}

// ReturnGame handles POST /borrowings/:id/return - return a borrowed game with HTMX response
func (h *BorrowingWebHandler) ReturnGame(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid borrowing ID",
		})
		return
	}

	// Get borrowing details before returning
	borrowing, err := h.borrowingService.GetBorrowingDetails(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "Borrowing not found",
		})
		return
	}

	if err := h.borrowingService.ReturnGame(id); err != nil {
		c.HTML(http.StatusConflict, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to return game: " + err.Error(),
		})
		return
	}

	// Trigger multiple updates
	c.Header("HX-Trigger", "game-returned, refresh-games, refresh-dashboard, refresh-alerts")
	c.HTML(http.StatusOK, "borrowings/return-success.html", gin.H{
		"Message":   "Game returned successfully!",
		"Borrowing": borrowing,
	})
}

// ShowExtendForm handles GET /borrowings/:id/extend - display extend due date form
func (h *BorrowingWebHandler) ShowExtendForm(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid borrowing ID",
		})
		return
	}

	borrowing, err := h.borrowingService.GetBorrowingDetails(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "Borrowing not found",
		})
		return
	}

	c.HTML(http.StatusOK, "borrowings/extend.html", gin.H{
		"Borrowing":    borrowing,
		"MinDate":      time.Now().Format("2006-01-02"),
		"MaxDate":      borrowing.BorrowedAt.AddDate(0, 0, 90).Format("2006-01-02"),
		"SuggestedDate": borrowing.DueDate.AddDate(0, 0, 14).Format("2006-01-02"),
	})
}

// ExtendDueDate handles POST /borrowings/:id/extend - extend due date with HTMX response
func (h *BorrowingWebHandler) ExtendDueDate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid borrowing ID",
		})
		return
	}

	newDueDateStr := c.PostForm("new_due_date")
	newDueDate, err := time.Parse("2006-01-02", newDueDateStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid due date format",
		})
		return
	}

	if err := h.borrowingService.ExtendDueDate(id, newDueDate); err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Failed to extend due date: " + err.Error(),
		})
		return
	}

	// Get updated borrowing details
	borrowing, _ := h.borrowingService.GetBorrowingDetails(id)

	c.Header("HX-Trigger", "due-date-extended, refresh-dashboard, refresh-alerts")
	c.HTML(http.StatusOK, "borrowings/extend-success.html", gin.H{
		"Message":   "Due date extended successfully!",
		"Borrowing": borrowing,
	})
}

// ShowBorrowingDetails handles GET /borrowings/:id - display borrowing details modal
func (h *BorrowingWebHandler) ShowBorrowingDetails(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid borrowing ID",
		})
		return
	}

	borrowing, err := h.borrowingService.GetBorrowingDetails(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "Borrowing not found",
		})
		return
	}

	c.HTML(http.StatusOK, "borrowings/detail.html", gin.H{
		"Borrowing": borrowing,
	})
}

// GetGameAvailabilityStatus handles GET /games/:id/availability - return availability status for HTMX updates
func (h *BorrowingWebHandler) GetGameAvailabilityStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/error.html", gin.H{
			"ErrorMessage": "Invalid game ID",
		})
		return
	}

	game, err := h.gameService.GetGame(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/error.html", gin.H{
			"ErrorMessage": "Game not found",
		})
		return
	}

	c.HTML(http.StatusOK, "games/partials/availability-status.html", gin.H{
		"Game": game,
	})
}

// RegisterWebRoutes registers all borrowing web routes
func (h *BorrowingWebHandler) RegisterWebRoutes(router *gin.RouterGroup) {
	borrowings := router.Group("/borrowings")
	{
		borrowings.GET("/new", h.ShowNewBorrowingForm)
		borrowings.POST("", h.CreateBorrowing)
		borrowings.GET("/:id", h.ShowBorrowingDetails)
		borrowings.POST("/:id/return", h.ReturnGame)
		borrowings.GET("/:id/extend", h.ShowExtendForm)
		borrowings.POST("/:id/extend", h.ExtendDueDate)
	}

	// Game availability status endpoint
	router.GET("/games/:id/availability", h.GetGameAvailabilityStatus)
}