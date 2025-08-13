package handlers

import (
	"board-game-library/internal/models"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GameWebHandler handles web requests for game management with HTMX support
type GameWebHandler struct {
	gameService GameServiceInterface
}

// NewGameWebHandler creates a new GameWebHandler instance
func NewGameWebHandler(gameService GameServiceInterface) *GameWebHandler {
	return &GameWebHandler{
		gameService: gameService,
	}
}

// ListGames handles GET /games - display games list page
func (h *GameWebHandler) ListGames(c *gin.Context) {
	games, err := h.gameService.GetAllGames()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Title":        "Error",
			"ErrorMessage": "Failed to load games: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "games/list.html", gin.H{
		"Title":      "Games",
		"Games":      games,
		"TotalGames": len(games),
	})
}

// SearchGames handles POST /games/search - HTMX search for games
func (h *GameWebHandler) SearchGames(c *gin.Context) {
	query := strings.TrimSpace(c.PostForm("search"))
	
	var games []*models.Game
	var err error
	
	if query == "" {
		games, err = h.gameService.GetAllGames()
	} else {
		games, err = h.gameService.SearchGames(query)
	}
	
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Search failed: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "games/partials/games-grid.html", gin.H{
		"Games": games,
	})
}

// SearchFilterGames handles POST /games/search-filter - Combined HTMX search and filter
func (h *GameWebHandler) SearchFilterGames(c *gin.Context) {
	query := strings.TrimSpace(c.PostForm("search"))
	availability := c.PostForm("availability")
	
	var games []*models.Game
	var err error
	
	// First get games based on search query
	if query == "" {
		games, err = h.gameService.GetAllGames()
	} else {
		games, err = h.gameService.SearchGames(query)
	}
	
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Search failed: " + err.Error(),
		})
		return
	}
	
	// Then filter by availability
	if availability != "" && availability != "all" {
		filteredGames := make([]*models.Game, 0)
		for _, game := range games {
			switch availability {
			case "available":
				if game.IsAvailable {
					filteredGames = append(filteredGames, game)
				}
			case "borrowed":
				if !game.IsAvailable {
					filteredGames = append(filteredGames, game)
				}
			}
		}
		games = filteredGames
	}

	// Return both grid and table views
	c.HTML(http.StatusOK, "games/partials/games-container.html", gin.H{
		"Games": games,
	})
}

// FilterGames handles POST /games/filter - HTMX filter for games by availability
func (h *GameWebHandler) FilterGames(c *gin.Context) {
	availability := c.PostForm("availability")
	
	var games []*models.Game
	var err error
	
	switch availability {
	case "available":
		games, err = h.gameService.GetAvailableGames()
	case "borrowed":
		allGames, getAllErr := h.gameService.GetAllGames()
		if getAllErr != nil {
			err = getAllErr
		} else {
			games = make([]*models.Game, 0)
			for _, game := range allGames {
				if !game.IsAvailable {
					games = append(games, game)
				}
			}
		}
	default: // "all"
		games, err = h.gameService.GetAllGames()
	}
	
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Filter failed: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "games/partials/games-grid.html", gin.H{
		"Games": games,
	})
}

// SortGames handles POST /games/sort - HTMX sort for games
func (h *GameWebHandler) SortGames(c *gin.Context) {
	field := c.PostForm("field")
	direction := c.PostForm("direction")
	
	games, err := h.gameService.GetAllGames()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/error.html", gin.H{
			"ErrorMessage": "Sort failed: " + err.Error(),
		})
		return
	}

	// Sort games based on field and direction
	// This is a simple implementation - in a real app you'd want to do this in the service/repository layer
	sortGames(games, field, direction)

	c.HTML(http.StatusOK, "games/partials/games-table.html", gin.H{
		"Games":         games,
		"SortField":     field,
		"SortDirection": direction,
	})
}

// ShowGame handles GET /games/:id - display game details modal
func (h *GameWebHandler) ShowGame(c *gin.Context) {
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

	// Get borrowing history
	borrowings, _ := h.gameService.GetGameBorrowingHistory(id)

	c.HTML(http.StatusOK, "games/detail.html", gin.H{
		"Game":       game,
		"Borrowings": borrowings,
	})
}

// ShowNewGameForm handles GET /games/new - display new game form modal
func (h *GameWebHandler) ShowNewGameForm(c *gin.Context) {
	c.HTML(http.StatusOK, "games/new.html", gin.H{
		"Title": "Add New Game",
	})
}

// ShowEditGameForm handles GET /games/:id/edit - display edit game form modal
func (h *GameWebHandler) ShowEditGameForm(c *gin.Context) {
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

	c.HTML(http.StatusOK, "games/edit.html", gin.H{
		"Game": game,
	})
}

// RegisterWebRoutes registers all game web routes
func (h *GameWebHandler) RegisterWebRoutes(router *gin.RouterGroup) {
	games := router.Group("/games")
	{
		games.GET("", h.ListGames)
		games.POST("/search", h.SearchGames)
		games.POST("/search-filter", h.SearchFilterGames)
		games.POST("/filter", h.FilterGames)
		games.POST("/sort", h.SortGames)
		games.GET("/new", h.ShowNewGameForm)
		games.GET("/:id", h.ShowGame)
		games.GET("/:id/edit", h.ShowEditGameForm)
	}
}

// sortGames sorts games slice based on field and direction
func sortGames(games []*models.Game, field, direction string) {
	switch field {
	case "name":
		sort.Slice(games, func(i, j int) bool {
			if direction == "desc" {
				return games[i].Name > games[j].Name
			}
			return games[i].Name < games[j].Name
		})
	case "entry_date":
		sort.Slice(games, func(i, j int) bool {
			if direction == "desc" {
				return games[i].EntryDate.After(games[j].EntryDate)
			}
			return games[i].EntryDate.Before(games[j].EntryDate)
		})
	case "category":
		sort.Slice(games, func(i, j int) bool {
			if direction == "desc" {
				return games[i].Category > games[j].Category
			}
			return games[i].Category < games[j].Category
		})
	}
}