package handlers

import (
	"board-game-library/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GameServiceInterface defines the interface for game service operations
type GameServiceInterface interface {
	AddGame(name, description, category, condition string) (*models.Game, error)
	GetGame(id int) (*models.Game, error)
	GetAllGames() ([]*models.Game, error)
	GetAvailableGames() ([]*models.Game, error)
	SearchGames(query string) ([]*models.Game, error)
	UpdateGame(game *models.Game) error
	SetGameAvailability(gameID int, isAvailable bool) error
	GetGameBorrowingHistory(gameID int) ([]*models.Borrowing, error)
	IsGameAvailable(gameID int) (bool, error)
	GetCurrentBorrower(gameID int) (*models.Borrowing, error)
	DeleteGame(gameID int) error
}

// GameHandler handles HTTP requests for game management
type GameHandler struct {
	gameService GameServiceInterface
}

// NewGameHandler creates a new GameHandler instance
func NewGameHandler(gameService GameServiceInterface) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

// AddGameRequest represents the request body for adding a new game
type AddGameRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Condition   string `json:"condition"`
}

// AddGame handles POST /api/games - add a new game
// @Summary Ajouter un nouveau jeu
// @Description Ajoute un nouveau jeu à la bibliothèque
// @Tags games
// @Accept json
// @Produce json
// @Param game body AddGameRequest true "Informations du jeu"
// @Success 201 {object} map[string]interface{} "Jeu créé avec succès"
// @Failure 400 {object} map[string]interface{} "Données invalides"
// @Failure 500 {object} map[string]interface{} "Erreur serveur"
// @Router /games [post]
func (h *GameHandler) AddGame(c *gin.Context) {
	var req AddGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Set default condition if not provided
	if req.Condition == "" {
		req.Condition = "good"
	}

	game, err := h.gameService.AddGame(req.Name, req.Description, req.Category, req.Condition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add game",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Game added successfully",
		"game":    game,
	})
}

// GetAllGames handles GET /api/games - list all games with optional search
// @Summary Lister tous les jeux
// @Description Récupère la liste de tous les jeux avec recherche optionnelle
// @Tags games
// @Accept json
// @Produce json
// @Param search query string false "Terme de recherche"
// @Param available query boolean false "Filtrer par disponibilité"
// @Success 200 {object} map[string]interface{} "Liste des jeux"
// @Failure 500 {object} map[string]interface{} "Erreur serveur"
// @Router /games [get]
func (h *GameHandler) GetAllGames(c *gin.Context) {
	query := c.Query("search")
	availableOnly := c.Query("available") == "true"

	var games []*models.Game
	var err error

	if availableOnly {
		games, err = h.gameService.GetAvailableGames()
	} else if query != "" {
		games, err = h.gameService.SearchGames(query)
	} else {
		games, err = h.gameService.GetAllGames()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve games",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"games": games,
		"count": len(games),
	})
}

// GetGame handles GET /api/games/:id - get game by ID
func (h *GameHandler) GetGame(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid game ID",
			"details": "Game ID must be a valid integer",
		})
		return
	}

	game, err := h.gameService.GetGame(id)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Game not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve game",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game": game,
	})
}

// UpdateGameRequest represents the request body for updating a game
type UpdateGameRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Condition   string `json:"condition"`
	IsAvailable *bool  `json:"is_available"`
}

// UpdateGame handles PUT /api/games/:id - update game information
func (h *GameHandler) UpdateGame(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid game ID",
			"details": "Game ID must be a valid integer",
		})
		return
	}

	var req UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get existing game
	existingGame, err := h.gameService.GetGame(id)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Game not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve game",
			"details": err.Error(),
		})
		return
	}

	// Update game fields
	existingGame.Name = req.Name
	existingGame.Description = req.Description
	existingGame.Category = req.Category
	existingGame.Condition = req.Condition
	if req.IsAvailable != nil {
		existingGame.IsAvailable = *req.IsAvailable
	}

	// Update game
	if err := h.gameService.UpdateGame(existingGame); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update game",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game updated successfully",
		"game":    existingGame,
	})
}

// DeleteGame handles DELETE /api/games/:id - delete a game
func (h *GameHandler) DeleteGame(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid game ID",
			"details": "Game ID must be a valid integer",
		})
		return
	}

	if err := h.gameService.DeleteGame(id); err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Game not found",
				"details": err.Error(),
			})
			return
		}
		if err.Error() == "cannot delete game: currently borrowed" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Cannot delete game",
				"details": "Game is currently borrowed and cannot be deleted",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete game",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game deleted successfully",
	})
}

// GetGameBorrowingHistory handles GET /api/games/:id/borrowings - get game borrowing history
func (h *GameHandler) GetGameBorrowingHistory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid game ID",
			"details": "Game ID must be a valid integer",
		})
		return
	}

	borrowings, err := h.gameService.GetGameBorrowingHistory(id)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Game not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve game borrowing history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"borrowings": borrowings,
		"count":      len(borrowings),
	})
}

// GetGameAvailability handles GET /api/games/:id/availability - check game availability
func (h *GameHandler) GetGameAvailability(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid game ID",
			"details": "Game ID must be a valid integer",
		})
		return
	}

	isAvailable, err := h.gameService.IsGameAvailable(id)
	if err != nil {
		if err.Error() == "game not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Game not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check game availability",
			"details": err.Error(),
		})
		return
	}

	response := gin.H{
		"is_available": isAvailable,
	}

	// If not available, get current borrower info
	if !isAvailable {
		currentBorrower, err := h.gameService.GetCurrentBorrower(id)
		if err == nil && currentBorrower != nil {
			response["current_borrower"] = currentBorrower
		}
	}

	c.JSON(http.StatusOK, response)
}

// SearchGames handles GET /api/games/search - search games with query parameters
func (h *GameHandler) SearchGames(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing search query",
			"details": "Query parameter 'q' is required",
		})
		return
	}

	games, err := h.gameService.SearchGames(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search games",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"games": games,
		"count": len(games),
		"query": query,
	})
}

// RegisterRoutes registers all game-related routes
func (h *GameHandler) RegisterRoutes(router *gin.RouterGroup) {
	games := router.Group("/games")
	{
		games.POST("", h.AddGame)
		games.GET("", h.GetAllGames)
		games.GET("/search", h.SearchGames)
		games.GET("/:id", h.GetGame)
		games.PUT("/:id", h.UpdateGame)
		games.DELETE("/:id", h.DeleteGame)
		games.GET("/:id/borrowings", h.GetGameBorrowingHistory)
		games.GET("/:id/availability", h.GetGameAvailability)
	}
}