package repositories

import (
	"board-game-library/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id int) error
	GetBorrowingHistory(userID int) ([]*models.Borrowing, error)
}

// GameRepository defines the interface for game data operations
type GameRepository interface {
	Create(game *models.Game) error
	GetByID(id int) (*models.Game, error)
	GetAll() ([]*models.Game, error)
	Search(query string) ([]*models.Game, error)
	Update(game *models.Game) error
	Delete(id int) error
	GetAvailable() ([]*models.Game, error)
}

// BorrowingRepository defines the interface for borrowing data operations
type BorrowingRepository interface {
	Create(borrowing *models.Borrowing) error
	GetByID(id int) (*models.Borrowing, error)
	GetActiveByUser(userID int) ([]*models.Borrowing, error)
	GetByGame(gameID int) ([]*models.Borrowing, error)
	GetOverdue() ([]*models.Borrowing, error)
	Update(borrowing *models.Borrowing) error
	ReturnGame(borrowingID int) error
	GetAll() ([]*models.Borrowing, error)
}

// AlertRepository defines the interface for alert data operations
type AlertRepository interface {
	Create(alert *models.Alert) error
	GetByID(id int) (*models.Alert, error)
	GetUnread() ([]*models.Alert, error)
	GetByUser(userID int) ([]*models.Alert, error)
	GetAll() ([]*models.Alert, error)
	MarkAsRead(id int) error
	Delete(id int) error
}