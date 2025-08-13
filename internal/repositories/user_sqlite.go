package repositories

import (
	"board-game-library/internal/models"
	"board-game-library/pkg/database"
	"database/sql"
	"fmt"
)

// SQLiteUserRepository implements UserRepository using SQLite
type SQLiteUserRepository struct {
	db *database.DB
}

// NewSQLiteUserRepository creates a new SQLite user repository
func NewSQLiteUserRepository(db *database.DB) UserRepository {
	return &SQLiteUserRepository{db: db}
}

// Create inserts a new user into the database
func (r *SQLiteUserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (name, email, registered_at, is_active)
		VALUES (?, ?, ?, ?)
		RETURNING id`
	
	err := r.db.QueryRow(query, user.Name, user.Email, user.RegisteredAt, user.IsActive).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetByID retrieves a user by their ID
func (r *SQLiteUserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, name, email, registered_at, is_active
		FROM users
		WHERE id = ?`
	
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.RegisteredAt, &user.IsActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	
	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *SQLiteUserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, name, email, registered_at, is_active
		FROM users
		WHERE email = ?`
	
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.RegisteredAt, &user.IsActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

// GetAll retrieves all users from the database
func (r *SQLiteUserRepository) GetAll() ([]*models.User, error) {
	query := `
		SELECT id, name, email, registered_at, is_active
		FROM users
		ORDER BY name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.RegisteredAt, &user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}
	
	return users, nil
}

// Update modifies an existing user in the database
func (r *SQLiteUserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET name = ?, email = ?, is_active = ?
		WHERE id = ?`
	
	result, err := r.db.Exec(query, user.Name, user.Email, user.IsActive, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", user.ID)
	}
	
	return nil
}

// Delete removes a user from the database
func (r *SQLiteUserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	
	return nil
}

// GetBorrowingHistory retrieves the borrowing history for a user
func (r *SQLiteUserRepository) GetBorrowingHistory(userID int) ([]*models.Borrowing, error) {
	query := `
		SELECT b.id, b.user_id, b.game_id, b.borrowed_at, b.due_date, b.returned_at, b.is_overdue
		FROM borrowings b
		WHERE b.user_id = ?
		ORDER BY b.borrowed_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get borrowing history: %w", err)
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