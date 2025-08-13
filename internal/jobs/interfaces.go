package jobs

// AlertService defines the interface for alert service operations needed by the job system
type AlertService interface {
	GenerateOverdueAlerts() error
	GenerateReminderAlerts() error
	CleanupResolvedAlerts() error
}