package services

import (
	"board-game-library/internal/models"
	"board-game-library/internal/repositories"
	"fmt"
	"time"
)

// AlertService handles alert-related business logic
type AlertService struct {
	alertRepo     repositories.AlertRepository
	borrowingRepo repositories.BorrowingRepository
	userRepo      repositories.UserRepository
	gameRepo      repositories.GameRepository
}

// NewAlertService creates a new AlertService instance
func NewAlertService(alertRepo repositories.AlertRepository, borrowingRepo repositories.BorrowingRepository, userRepo repositories.UserRepository, gameRepo repositories.GameRepository) *AlertService {
	return &AlertService{
		alertRepo:     alertRepo,
		borrowingRepo: borrowingRepo,
		userRepo:      userRepo,
		gameRepo:      gameRepo,
	}
}

// GenerateOverdueAlerts creates alerts for all overdue items
func (s *AlertService) GenerateOverdueAlerts() error {
	// Get all overdue borrowings
	overdueBorrowings, err := s.borrowingRepo.GetOverdue()
	if err != nil {
		return fmt.Errorf("failed to get overdue borrowings: %w", err)
	}

	// Filter to only include items that are actually overdue and not returned
	var activeOverdue []*models.Borrowing
	for _, borrowing := range overdueBorrowings {
		if borrowing.ReturnedAt == nil && borrowing.IsCurrentlyOverdue() {
			activeOverdue = append(activeOverdue, borrowing)
		}
	}

	// Create alerts for each overdue item
	for _, borrowing := range activeOverdue {
		// Check if alert already exists for this borrowing
		existingAlerts, err := s.alertRepo.GetByUser(borrowing.UserID)
		if err != nil {
			return fmt.Errorf("failed to check existing alerts for user %d: %w", borrowing.UserID, err)
		}

		// Check if we already have an overdue alert for this game
		hasExistingAlert := false
		for _, alert := range existingAlerts {
			if alert.GameID == borrowing.GameID && alert.Type == "overdue" && !alert.IsRead {
				hasExistingAlert = true
				break
			}
		}

		if !hasExistingAlert {
			// Get game details for the alert message
			game, err := s.gameRepo.GetByID(borrowing.GameID)
			if err != nil {
				return fmt.Errorf("failed to get game details for alert: %w", err)
			}

			// Calculate days overdue
			daysOverdue := borrowing.DaysOverdue()
			message := fmt.Sprintf("Game '%s' is overdue by %d day(s). Please return it as soon as possible.", game.Name, daysOverdue)

			// Create overdue alert
			alert := &models.Alert{
				UserID:    borrowing.UserID,
				GameID:    borrowing.GameID,
				Type:      "overdue",
				Message:   message,
				CreatedAt: time.Now(),
				IsRead:    false,
			}

			if err := models.ValidateAlert(alert); err != nil {
				return fmt.Errorf("alert validation failed: %w", err)
			}

			if err := s.alertRepo.Create(alert); err != nil {
				return fmt.Errorf("failed to create overdue alert: %w", err)
			}
		}
	}

	return nil
}

// GenerateReminderAlerts creates reminder alerts for items due within 2 days
func (s *AlertService) GenerateReminderAlerts() error {
	// Get all active borrowings
	allBorrowings, err := s.borrowingRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get all borrowings: %w", err)
	}

	// Filter items due within 2 days
	cutoffDate := time.Now().Add(2 * 24 * time.Hour)
	var itemsDueSoon []*models.Borrowing

	for _, borrowing := range allBorrowings {
		if borrowing.ReturnedAt == nil && // Not returned
			borrowing.DueDate.Before(cutoffDate) && // Due within 2 days
			!borrowing.IsCurrentlyOverdue() { // Not already overdue
			itemsDueSoon = append(itemsDueSoon, borrowing)
		}
	}

	// Create reminder alerts for each item due soon
	for _, borrowing := range itemsDueSoon {
		// Check if reminder alert already exists for this borrowing
		existingAlerts, err := s.alertRepo.GetByUser(borrowing.UserID)
		if err != nil {
			return fmt.Errorf("failed to check existing alerts for user %d: %w", borrowing.UserID, err)
		}

		// Check if we already have a reminder alert for this game
		hasExistingAlert := false
		for _, alert := range existingAlerts {
			if alert.GameID == borrowing.GameID && alert.Type == "reminder" && !alert.IsRead {
				hasExistingAlert = true
				break
			}
		}

		if !hasExistingAlert {
			// Get game details for the alert message
			game, err := s.gameRepo.GetByID(borrowing.GameID)
			if err != nil {
				return fmt.Errorf("failed to get game details for alert: %w", err)
			}

			// Calculate days until due
			daysUntilDue := int(time.Until(borrowing.DueDate).Hours() / 24)
			if daysUntilDue < 0 {
				daysUntilDue = 0
			}

			var message string
			if daysUntilDue == 0 {
				message = fmt.Sprintf("Game '%s' is due today. Please return it by the end of the day.", game.Name)
			} else {
				message = fmt.Sprintf("Game '%s' is due in %d day(s). Please plan to return it soon.", game.Name, daysUntilDue)
			}

			// Create reminder alert
			alert := &models.Alert{
				UserID:    borrowing.UserID,
				GameID:    borrowing.GameID,
				Type:      "reminder",
				Message:   message,
				CreatedAt: time.Now(),
				IsRead:    false,
			}

			if err := models.ValidateAlert(alert); err != nil {
				return fmt.Errorf("alert validation failed: %w", err)
			}

			if err := s.alertRepo.Create(alert); err != nil {
				return fmt.Errorf("failed to create reminder alert: %w", err)
			}
		}
	}

	return nil
}

// GetActiveAlerts retrieves all unread alerts
func (s *AlertService) GetActiveAlerts() ([]*models.Alert, error) {
	alerts, err := s.alertRepo.GetUnread()
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}

	return alerts, nil
}

// GetAlertsByUser retrieves all alerts for a specific user
func (s *AlertService) GetAlertsByUser(userID int) ([]*models.Alert, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	alerts, err := s.alertRepo.GetByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user alerts: %w", err)
	}

	return alerts, nil
}

// MarkAlertAsRead marks a specific alert as read
func (s *AlertService) MarkAlertAsRead(alertID int) error {
	if alertID <= 0 {
		return fmt.Errorf("invalid alert ID: %d", alertID)
	}

	// Verify alert exists
	_, err := s.alertRepo.GetByID(alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	if err := s.alertRepo.MarkAsRead(alertID); err != nil {
		return fmt.Errorf("failed to mark alert as read: %w", err)
	}

	return nil
}

// MarkAllUserAlertsAsRead marks all alerts for a user as read
func (s *AlertService) MarkAllUserAlertsAsRead(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID: %d", userID)
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Get all user alerts
	alerts, err := s.alertRepo.GetByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to get user alerts: %w", err)
	}

	// Mark each unread alert as read
	for _, alert := range alerts {
		if !alert.IsRead {
			if err := s.alertRepo.MarkAsRead(alert.ID); err != nil {
				return fmt.Errorf("failed to mark alert %d as read: %w", alert.ID, err)
			}
		}
	}

	return nil
}

// DeleteAlert removes an alert
func (s *AlertService) DeleteAlert(alertID int) error {
	if alertID <= 0 {
		return fmt.Errorf("invalid alert ID: %d", alertID)
	}

	// Verify alert exists
	_, err := s.alertRepo.GetByID(alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	if err := s.alertRepo.Delete(alertID); err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	return nil
}

// CleanupResolvedAlerts removes alerts for items that have been returned
func (s *AlertService) CleanupResolvedAlerts() error {
	// Get all alerts
	allAlerts, err := s.alertRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get all alerts: %w", err)
	}

	// Check each alert to see if the associated borrowing has been resolved
	for _, alert := range allAlerts {
		// Get borrowings for this game by this user
		borrowings, err := s.borrowingRepo.GetByGame(alert.GameID)
		if err != nil {
			continue // Skip this alert if we can't check borrowings
		}

		// Check if there's an active borrowing for this user and game
		hasActiveBorrowing := false
		for _, borrowing := range borrowings {
			if borrowing.UserID == alert.UserID && borrowing.ReturnedAt == nil {
				hasActiveBorrowing = true
				break
			}
		}

		// If no active borrowing, delete the alert
		if !hasActiveBorrowing {
			if err := s.alertRepo.Delete(alert.ID); err != nil {
				return fmt.Errorf("failed to delete resolved alert %d: %w", alert.ID, err)
			}
		}
	}

	return nil
}

// GetAlertsSummaryByUser returns a summary of alerts grouped by user
func (s *AlertService) GetAlertsSummaryByUser() (map[int]AlertSummary, error) {
	// Get all unread alerts
	alerts, err := s.alertRepo.GetUnread()
	if err != nil {
		return nil, fmt.Errorf("failed to get unread alerts: %w", err)
	}

	// Group alerts by user
	summary := make(map[int]AlertSummary)
	for _, alert := range alerts {
		userSummary := summary[alert.UserID]
		userSummary.UserID = alert.UserID
		userSummary.TotalAlerts++

		if alert.Type == "overdue" {
			userSummary.OverdueCount++
		} else if alert.Type == "reminder" {
			userSummary.ReminderCount++
		}

		userSummary.Alerts = append(userSummary.Alerts, alert)
		summary[alert.UserID] = userSummary
	}

	return summary, nil
}

// AlertSummary represents a summary of alerts for a user
type AlertSummary struct {
	UserID        int             `json:"user_id"`
	TotalAlerts   int             `json:"total_alerts"`
	OverdueCount  int             `json:"overdue_count"`
	ReminderCount int             `json:"reminder_count"`
	Alerts        []*models.Alert `json:"alerts"`
}

// CreateCustomAlert creates a custom alert for a user
func (s *AlertService) CreateCustomAlert(userID, gameID int, alertType, message string) (*models.Alert, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}
	if gameID <= 0 {
		return nil, fmt.Errorf("invalid game ID: %d", gameID)
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify game exists
	_, err = s.gameRepo.GetByID(gameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// Create alert
	alert := &models.Alert{
		UserID:    userID,
		GameID:    gameID,
		Type:      alertType,
		Message:   message,
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	// Validate alert
	if err := models.ValidateAlert(alert); err != nil {
		return nil, fmt.Errorf("alert validation failed: %w", err)
	}

	// Create alert in repository
	if err := s.alertRepo.Create(alert); err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	return alert, nil
}