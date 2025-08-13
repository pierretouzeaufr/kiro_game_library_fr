package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager manages the job scheduler lifecycle and provides a high-level interface
type Manager struct {
	scheduler    *Scheduler
	alertService AlertService
	logger       *log.Logger
	started      bool
	mu           sync.RWMutex
}

// Config holds configuration for the job manager
type Config struct {
	// EnableOverdueAlerts enables the daily overdue alert generation job
	EnableOverdueAlerts bool
	
	// EnableReminderAlerts enables the daily reminder alert generation job
	EnableReminderAlerts bool
	
	// EnableAlertCleanup enables the periodic alert cleanup job
	EnableAlertCleanup bool
	
	// OverdueAlertSchedule sets the schedule for overdue alert generation (default: 24h)
	OverdueAlertSchedule time.Duration
	
	// ReminderAlertSchedule sets the schedule for reminder alert generation (default: 24h)
	ReminderAlertSchedule time.Duration
	
	// CleanupSchedule sets the schedule for alert cleanup (default: 6h)
	CleanupSchedule time.Duration
	
	// Logger for job operations
	Logger *log.Logger
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		EnableOverdueAlerts:   true,
		EnableReminderAlerts:  true,
		EnableAlertCleanup:    true,
		OverdueAlertSchedule:  24 * time.Hour,
		ReminderAlertSchedule: 24 * time.Hour,
		CleanupSchedule:       6 * time.Hour,
		Logger:                log.New(log.Writer(), "[JOB-MANAGER] ", log.LstdFlags),
	}
}

// NewManager creates a new job manager
func NewManager(alertService AlertService, config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}
	
	if config.Logger == nil {
		config.Logger = log.New(log.Writer(), "[JOB-MANAGER] ", log.LstdFlags)
	}
	
	scheduler := NewScheduler(alertService, config.Logger)
	
	manager := &Manager{
		scheduler:    scheduler,
		alertService: alertService,
		logger:       config.Logger,
		started:      false,
	}
	
	// Configure jobs based on config
	manager.configureJobs(config)
	
	return manager
}

// configureJobs configures the scheduler based on the provided configuration
func (m *Manager) configureJobs(config *Config) {
	// Remove default jobs first
	m.scheduler.RemoveJob("overdue-alerts")
	m.scheduler.RemoveJob("reminder-alerts")
	m.scheduler.RemoveJob("cleanup-alerts")
	
	// Add jobs based on configuration
	if config.EnableOverdueAlerts {
		m.scheduler.AddJob("overdue-alerts", "Generate alerts for overdue items", config.OverdueAlertSchedule, func() error {
			m.logger.Println("Starting overdue alert generation")
			err := m.alertService.GenerateOverdueAlerts()
			if err != nil {
				m.logger.Printf("Failed to generate overdue alerts: %v", err)
				return fmt.Errorf("overdue alert generation failed: %w", err)
			}
			m.logger.Println("Overdue alert generation completed successfully")
			return nil
		})
	}
	
	if config.EnableReminderAlerts {
		m.scheduler.AddJob("reminder-alerts", "Generate reminder alerts for items due soon", config.ReminderAlertSchedule, func() error {
			m.logger.Println("Starting reminder alert generation")
			err := m.alertService.GenerateReminderAlerts()
			if err != nil {
				m.logger.Printf("Failed to generate reminder alerts: %v", err)
				return fmt.Errorf("reminder alert generation failed: %w", err)
			}
			m.logger.Println("Reminder alert generation completed successfully")
			return nil
		})
	}
	
	if config.EnableAlertCleanup {
		m.scheduler.AddJob("cleanup-alerts", "Clean up alerts for returned items", config.CleanupSchedule, func() error {
			m.logger.Println("Starting alert cleanup")
			err := m.alertService.CleanupResolvedAlerts()
			if err != nil {
				m.logger.Printf("Failed to cleanup resolved alerts: %v", err)
				return fmt.Errorf("alert cleanup failed: %w", err)
			}
			m.logger.Println("Alert cleanup completed successfully")
			return nil
		})
	}
}

// Start starts the job manager and scheduler
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.started {
		return fmt.Errorf("job manager is already started")
	}
	
	m.logger.Println("Starting job manager")
	m.scheduler.Start()
	m.started = true
	
	// Start a goroutine to handle context cancellation
	go func() {
		<-ctx.Done()
		m.Stop()
	}()
	
	m.logger.Println("Job manager started successfully")
	return nil
}

// Stop stops the job manager and scheduler
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.started {
		return fmt.Errorf("job manager is not started")
	}
	
	m.logger.Println("Stopping job manager")
	m.scheduler.Stop()
	m.started = false
	m.logger.Println("Job manager stopped successfully")
	
	return nil
}

// IsStarted returns whether the job manager is currently started
func (m *Manager) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

// RunJobNow executes a job immediately
func (m *Manager) RunJobNow(jobName string) error {
	return m.scheduler.RunJobNow(jobName)
}

// GetJobs returns all registered jobs
func (m *Manager) GetJobs() map[string]*Job {
	return m.scheduler.GetJobs()
}

// GetJobExecutions returns recent job executions
func (m *Manager) GetJobExecutions(limit int) []JobExecution {
	return m.scheduler.GetJobExecutions(limit)
}

// GetJobStatus returns the status of a specific job
func (m *Manager) GetJobStatus(jobName string) (*Job, error) {
	return m.scheduler.GetJobStatus(jobName)
}

// EnableJob enables a job
func (m *Manager) EnableJob(jobName string) error {
	return m.scheduler.EnableJob(jobName)
}

// DisableJob disables a job
func (m *Manager) DisableJob(jobName string) error {
	return m.scheduler.DisableJob(jobName)
}

// AddCustomJob adds a custom job to the scheduler
func (m *Manager) AddCustomJob(name, description string, schedule time.Duration, handler func() error) {
	m.scheduler.AddJob(name, description, schedule, handler)
}

// RemoveJob removes a job from the scheduler
func (m *Manager) RemoveJob(jobName string) {
	m.scheduler.RemoveJob(jobName)
}

// GetHealthStatus returns the health status of the job manager
func (m *Manager) GetHealthStatus() HealthStatus {
	m.mu.RLock()
	started := m.started
	m.mu.RUnlock()
	
	status := HealthStatus{
		IsRunning:     started,
		TotalJobs:     0,
		EnabledJobs:   0,
		DisabledJobs:  0,
		LastExecution: nil,
	}
	
	jobs := m.scheduler.GetJobs()
	status.TotalJobs = len(jobs)
	
	for _, job := range jobs {
		if job.Enabled {
			status.EnabledJobs++
		} else {
			status.DisabledJobs++
		}
	}
	
	// Get the most recent execution
	executions := m.scheduler.GetJobExecutions(1)
	if len(executions) > 0 {
		status.LastExecution = &executions[0]
	}
	
	return status
}

// HealthStatus represents the health status of the job manager
type HealthStatus struct {
	IsRunning     bool          `json:"is_running"`
	TotalJobs     int           `json:"total_jobs"`
	EnabledJobs   int           `json:"enabled_jobs"`
	DisabledJobs  int           `json:"disabled_jobs"`
	LastExecution *JobExecution `json:"last_execution,omitempty"`
}

// GenerateAllAlerts runs all alert generation jobs immediately
func (m *Manager) GenerateAllAlerts() error {
	jobs := []string{"overdue-alerts", "reminder-alerts"}
	
	for _, jobName := range jobs {
		if err := m.RunJobNow(jobName); err != nil {
			return fmt.Errorf("failed to run job %s: %w", jobName, err)
		}
	}
	
	return nil
}

// CleanupAlerts runs the alert cleanup job immediately
func (m *Manager) CleanupAlerts() error {
	return m.RunJobNow("cleanup-alerts")
}

// GetJobStatistics returns statistics about job executions
func (m *Manager) GetJobStatistics() JobStatistics {
	executions := m.scheduler.GetJobExecutions(0) // Get all executions
	
	stats := JobStatistics{
		TotalExecutions:    len(executions),
		SuccessfulExecutions: 0,
		FailedExecutions:   0,
		JobExecutionCounts: make(map[string]int),
		JobSuccessRates:    make(map[string]float64),
	}
	
	jobStats := make(map[string]struct {
		total   int
		success int
	})
	
	for _, exec := range executions {
		stats.JobExecutionCounts[exec.JobName]++
		
		jobStat := jobStats[exec.JobName]
		jobStat.total++
		
		if exec.Status == JobStatusCompleted {
			stats.SuccessfulExecutions++
			jobStat.success++
		} else if exec.Status == JobStatusFailed {
			stats.FailedExecutions++
		}
		
		jobStats[exec.JobName] = jobStat
	}
	
	// Calculate success rates
	for jobName, stat := range jobStats {
		if stat.total > 0 {
			stats.JobSuccessRates[jobName] = float64(stat.success) / float64(stat.total) * 100
		}
	}
	
	return stats
}

// JobStatistics represents statistics about job executions
type JobStatistics struct {
	TotalExecutions      int                `json:"total_executions"`
	SuccessfulExecutions int                `json:"successful_executions"`
	FailedExecutions     int                `json:"failed_executions"`
	JobExecutionCounts   map[string]int     `json:"job_execution_counts"`
	JobSuccessRates      map[string]float64 `json:"job_success_rates"`
}