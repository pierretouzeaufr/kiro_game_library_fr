package repositories

import (
	"board-game-library/internal/models"
	"board-game-library/pkg/database"
	"database/sql"
	"fmt"
	"time"
)

// SQLiteBorrowingRepository implements BorrowingRepository using SQLite
type SQLiteBorrowingRepository struct {
	db *database.DB
}

// NewSQLiteBorrowingRepository creates a new SQLite borrowing repository
func NewSQLiteBorrowingRepository(db *database.DB) BorrowingRepository {
	return &SQLiteBorrowingRepository{db: db}
}

// Create inserts a new borrowing record into the database
func (r *SQLiteBorrowingRepository) Create(borrowing *models.Borrowing) error {
	query := `
		INSERT INTO borrowings (user_id, game_id, borrowed_at, due_date, returned_at, is_overdue)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`
	
	err := r.db.QueryRow(query, borrowing.UserID, borrowing.GameID, borrowing.BorrowedAt,
		borrowing.DueDate, borrowing.ReturnedAt, borrowing.IsOverdue).Scan(&borrowing.ID)
	if err != nil {
		return fmt.Errorf("failed to create borrowing: %w", err)
	}
	
	return nil
}

// GetByID retrieves a borrowing record by its ID
func (r *SQLiteBorrowingRepository) GetByID(id int) (*models.Borrowing, error) {
	query := `
		SELECT id, user_id, game_id, borrowed_at, due_date, returned_at, is_overdue
		FROM borrowings
		WHERE id = ?`
	
	borrowing := &models.Borrowing{}
	err := r.db.QueryRow(query, id).Scan(
		&borrowing.ID, &borrowing.UserID, &borrowing.GameID,
		&borrowing.BorrowedAt, &borrowing.DueDate, &borrowing.ReturnedAt, &borrowing.IsOverdue,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("borrowing with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get borrowing by id: %w", err)
	}
	
	return borrowing, nil
}

// GetActiveByUser retrieves all active borrowings for a user
func (r *SQLiteBorrowingRepository) GetActiveByUser(userID int) ([]*models.Borrowing, error) {
	query := `
		SELECT id, user_id, game_id, borrowed_at, due_date, returned_at, is_overdue
		FROM borrowings
		WHERE user_id = ? AND returned_at IS NULL
		ORDER BY borrowed_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active borrowings by user: %w", err)
	}
	defer rows.Close()
	
	var borrowings []*models.Borrowing
	for rows.Next() {
		borrowing := &models.Borrowing{}
		err := rows.Scan(
			&borrowing.ID, &borrowing.UserID, &borrowing.GameID,
			&borrowing.BorrowedAt, &borrowing.DueDate, &borrowing.ReturnedAt, &borrowing.IsOverdue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan borrowing: %w", err)
		}
		borrowings = append(borrowings, borrowing)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating borrowings: %w", err)
	}
	
	return borrowings, nil
}

// GetByGame retrieves all borrowings for a specific game
func (r *SQLiteBorrowingRepository) GetByGame(gameID int) ([]*models.Borrowing, error) {
	query := `
		SELECT id, user_id, game_id, borrowed_at, due_date, returned_at, is_overdue
		FROM borrowings
		WHERE game_id = ?
		ORDER BY borrowed_at DESC`
	
	rows, err := r.db.Query(query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get borrowings by game: %w", err)
	}
	defer rows.Close()
	
	var borrowings []*models.Borrowing
	for rows.Next() {
		borrowing := &models.Borrowing{}
		err := rows.Scan(
			&borrowing.ID, &borrowing.UserID, &borrowing.GameID,
			&borrowing.BorrowedAt, &borrowing.DueDate, &borrowing.ReturnedAt, &borrowing.IsOverdue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan borrowing: %w", err)
		}
		borrowings = append(borrowings, borrowing)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating borrowings: %w", err)
	}
	
	return borrowings, nil
}

// GetOverdue retrieves all overdue borrowings
func (r *SQLiteBorrowingRepository) GetOverdue() ([]*models.Borrowing, error) {
	query := `
		SELECT id, user_id, game_id, borrowed_at, due_date, returned_at, is_overdue
		FROM borrowings
		WHERE returned_at IS NULL AND due_date < ?
		ORDER BY due_date ASC`
	
	rows, err := r.db.Query(query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue borrowings: %w", err)
	}
	defer rows.Close()
	
	var borrowings []*models.Borrowing
	for rows.Next() {
		borrowing := &models.Borrowing{}
		err := rows.Scan(
			&borrowing.ID, &borrowing.UserID, &borrowing.GameID,
			&borrowing.BorrowedAt, &borrowing.DueDate, &borrowing.ReturnedAt, &borrowing.IsOverdue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan borrowing: %w", err)
		}
		borrowings = append(borrowings, borrowing)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating borrowings: %w", err)
	}
	
	return borrowings, nil
}

// Update modifies an existing borrowing record
func (r *SQLiteBorrowingRepository) Update(borrowing *models.Borrowing) error {
	query := `
		UPDATE borrowings
		SET user_id = ?, game_id = ?, borrowed_at = ?, due_date = ?, returned_at = ?, is_overdue = ?
		WHERE id = ?`
	
	result, err := r.db.Exec(query, borrowing.UserID, borrowing.GameID, borrowing.BorrowedAt,
		borrowing.DueDate, borrowing.ReturnedAt, borrowing.IsOverdue, borrowing.ID)
	if err != nil {
		return fmt.Errorf("failed to update borrowing: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("borrowing with id %d not found", borrowing.ID)
	}
	
	return nil
}

// ReturnGame marks a borrowing as returned
func (r *SQLiteBorrowingRepository) ReturnGame(borrowingID int) error {
	query := `
		UPDATE borrowings
		SET returned_at = ?
		WHERE id = ? AND returned_at IS NULL`
	
	result, err := r.db.Exec(query, time.Now(), borrowingID)
	if err != nil {
		return fmt.Errorf("failed to return game: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("borrowing with id %d not found or already returned", borrowingID)
	}
	
	return nil
}

// GetAll retrieves all borrowing records
func (r *SQLiteBorrowingRepository) GetAll() ([]*models.Borrowing, error) {
	query := `
		SELECT id, user_id, game_id, borrowed_at, due_date, returned_at, is_overdue
		FROM borrowings
		ORDER BY borrowed_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all borrowings: %w", err)
	}
	defer rows.Close()
	
	var borrowings []*models.Borrowing
	for rows.Next() {
		borrowing := &models.Borrowing{}
		err := rows.Scan(
			&borrowing.ID, &borrowing.UserID, &borrowing.GameID,
			&borrowing.BorrowedAt, &borrowing.DueDate, &borrowing.ReturnedAt, &borrowing.IsOverdue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan borrowing: %w", err)
		}
		borrowings = append(borrowings, borrowing)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating borrowings: %w", err)
	}
	
	return borrowings, nil
}