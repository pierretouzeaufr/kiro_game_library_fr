package repositories

import (
	"board-game-library/internal/models"
	"board-game-library/pkg/database"
	"database/sql"
	"fmt"
)

// SQLiteAlertRepository implements AlertRepository using SQLite
type SQLiteAlertRepository struct {
	db *database.DB
}

// NewSQLiteAlertRepository creates a new SQLite alert repository
func NewSQLiteAlertRepository(db *database.DB) AlertRepository {
	return &SQLiteAlertRepository{db: db}
}

// Create inserts a new alert into the database
func (r *SQLiteAlertRepository) Create(alert *models.Alert) error {
	query := `
		INSERT INTO alerts (user_id, game_id, type, message, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`
	
	err := r.db.QueryRow(query, alert.UserID, alert.GameID, alert.Type,
		alert.Message, alert.CreatedAt, alert.IsRead).Scan(&alert.ID)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}
	
	return nil
}

// GetByID retrieves an alert by its ID
func (r *SQLiteAlertRepository) GetByID(id int) (*models.Alert, error) {
	query := `
		SELECT id, user_id, game_id, type, message, created_at, is_read
		FROM alerts
		WHERE id = ?`
	
	alert := &models.Alert{}
	err := r.db.QueryRow(query, id).Scan(
		&alert.ID, &alert.UserID, &alert.GameID, &alert.Type,
		&alert.Message, &alert.CreatedAt, &alert.IsRead,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get alert by id: %w", err)
	}
	
	return alert, nil
}

// GetUnread retrieves all unread alerts
func (r *SQLiteAlertRepository) GetUnread() ([]*models.Alert, error) {
	query := `
		SELECT id, user_id, game_id, type, message, created_at, is_read
		FROM alerts
		WHERE is_read = FALSE
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread alerts: %w", err)
	}
	defer rows.Close()
	
	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.GameID, &alert.Type,
			&alert.Message, &alert.CreatedAt, &alert.IsRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alerts: %w", err)
	}
	
	return alerts, nil
}

// GetByUser retrieves all alerts for a specific user
func (r *SQLiteAlertRepository) GetByUser(userID int) ([]*models.Alert, error) {
	query := `
		SELECT id, user_id, game_id, type, message, created_at, is_read
		FROM alerts
		WHERE user_id = ?
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts by user: %w", err)
	}
	defer rows.Close()
	
	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.GameID, &alert.Type,
			&alert.Message, &alert.CreatedAt, &alert.IsRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alerts: %w", err)
	}
	
	return alerts, nil
}

// GetAll retrieves all alerts from the database
func (r *SQLiteAlertRepository) GetAll() ([]*models.Alert, error) {
	query := `
		SELECT id, user_id, game_id, type, message, created_at, is_read
		FROM alerts
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all alerts: %w", err)
	}
	defer rows.Close()
	
	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.GameID, &alert.Type,
			&alert.Message, &alert.CreatedAt, &alert.IsRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alerts: %w", err)
	}
	
	return alerts, nil
}

// MarkAsRead marks an alert as read
func (r *SQLiteAlertRepository) MarkAsRead(id int) error {
	query := `
		UPDATE alerts
		SET is_read = TRUE
		WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark alert as read: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("alert with id %d not found", id)
	}
	
	return nil
}

// Delete removes an alert from the database
func (r *SQLiteAlertRepository) Delete(id int) error {
	query := `DELETE FROM alerts WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("alert with id %d not found", id)
	}
	
	return nil
}