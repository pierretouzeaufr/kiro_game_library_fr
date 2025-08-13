package repositories

import (
	"board-game-library/internal/models"
	"board-game-library/pkg/database"
	"database/sql"
	"fmt"
	"strings"
)

// SQLiteGameRepository implements GameRepository using SQLite
type SQLiteGameRepository struct {
	db *database.DB
}

// NewSQLiteGameRepository creates a new SQLite game repository
func NewSQLiteGameRepository(db *database.DB) GameRepository {
	return &SQLiteGameRepository{db: db}
}

// Create inserts a new game into the database
func (r *SQLiteGameRepository) Create(game *models.Game) error {
	query := `
		INSERT INTO games (name, description, category, entry_date, condition, is_available)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`
	
	err := r.db.QueryRow(query, game.Name, game.Description, game.Category, 
		game.EntryDate, game.Condition, game.IsAvailable).Scan(&game.ID)
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}
	
	return nil
}

// GetByID retrieves a game by its ID
func (r *SQLiteGameRepository) GetByID(id int) (*models.Game, error) {
	query := `
		SELECT id, name, description, category, entry_date, condition, is_available
		FROM games
		WHERE id = ?`
	
	game := &models.Game{}
	err := r.db.QueryRow(query, id).Scan(
		&game.ID, &game.Name, &game.Description, &game.Category,
		&game.EntryDate, &game.Condition, &game.IsAvailable,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get game by id: %w", err)
	}
	
	return game, nil
}

// GetAll retrieves all games from the database
func (r *SQLiteGameRepository) GetAll() ([]*models.Game, error) {
	query := `
		SELECT id, name, description, category, entry_date, condition, is_available
		FROM games
		ORDER BY name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all games: %w", err)
	}
	defer rows.Close()
	
	var games []*models.Game
	for rows.Next() {
		game := &models.Game{}
		err := rows.Scan(
			&game.ID, &game.Name, &game.Description, &game.Category,
			&game.EntryDate, &game.Condition, &game.IsAvailable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}
	
	return games, nil
}

// Search finds games matching the query string
func (r *SQLiteGameRepository) Search(query string) ([]*models.Game, error) {
	searchQuery := `
		SELECT id, name, description, category, entry_date, condition, is_available
		FROM games
		WHERE LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?
		ORDER BY name`
	
	searchTerm := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.Query(searchQuery, searchTerm, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search games: %w", err)
	}
	defer rows.Close()
	
	var games []*models.Game
	for rows.Next() {
		game := &models.Game{}
		err := rows.Scan(
			&game.ID, &game.Name, &game.Description, &game.Category,
			&game.EntryDate, &game.Condition, &game.IsAvailable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}
	
	return games, nil
}

// Update modifies an existing game in the database
func (r *SQLiteGameRepository) Update(game *models.Game) error {
	query := `
		UPDATE games
		SET name = ?, description = ?, category = ?, condition = ?, is_available = ?
		WHERE id = ?`
	
	result, err := r.db.Exec(query, game.Name, game.Description, game.Category,
		game.Condition, game.IsAvailable, game.ID)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("game with id %d not found", game.ID)
	}
	
	return nil
}

// Delete removes a game from the database
func (r *SQLiteGameRepository) Delete(id int) error {
	query := `DELETE FROM games WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("game with id %d not found", id)
	}
	
	return nil
}

// GetAvailable retrieves all available games
func (r *SQLiteGameRepository) GetAvailable() ([]*models.Game, error) {
	query := `
		SELECT id, name, description, category, entry_date, condition, is_available
		FROM games
		WHERE is_available = TRUE
		ORDER BY name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get available games: %w", err)
	}
	defer rows.Close()
	
	var games []*models.Game
	for rows.Next() {
		game := &models.Game{}
		err := rows.Scan(
			&game.ID, &game.Name, &game.Description, &game.Category,
			&game.EntryDate, &game.Condition, &game.IsAvailable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}
	
	return games, nil
}