package jobs

import (
	"board-game-library/internal/services"
	"context"
	"log"
	"os"
	"time"
)

// ExampleIntegration demonstrates how to integrate the job system with the main application
func ExampleIntegration() {
	// This is an example of how to integrate the job system with your main application
	// You would typically do this in your main.go or application startup code

	// 1. Create your alert service (assuming you have repositories set up)
	// alertService := services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	// 2. Create job manager configuration
	config := &Config{
		EnableOverdueAlerts:   true,
		EnableReminderAlerts:  true,
		EnableAlertCleanup:    true,
		OverdueAlertSchedule:  24 * time.Hour, // Run daily at the same time
		ReminderAlertSchedule: 24 * time.Hour, // Run daily at the same time
		CleanupSchedule:       6 * time.Hour,  // Run every 6 hours
		Logger:                log.New(os.Stdout, "[JOBS] ", log.LstdFlags),
	}

	// 3. Create job manager
	// manager := NewManager(alertService, config)

	// 4. Start the job manager with application context
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// err := manager.Start(ctx)
	// if err != nil {
	//     log.Fatalf("Failed to start job manager: %v", err)
	// }

	// 5. The job manager will now run in the background
	// You can also manually trigger jobs if needed:
	// manager.RunJobNow("overdue-alerts")
	// manager.GenerateAllAlerts()

	// 6. Get job status and health information
	// health := manager.GetHealthStatus()
	// log.Printf("Job manager health: %+v", health)

	// 7. When shutting down your application, stop the job manager
	// manager.Stop()
}

// IntegrateWithAlertService shows how to integrate with the actual alert service
func IntegrateWithAlertService(alertService *services.AlertService) *Manager {
	// Create a wrapper that implements our AlertService interface
	wrapper := &AlertServiceWrapper{alertService: alertService}
	
	// Create job manager with default configuration
	config := DefaultConfig()
	
	// Customize configuration as needed
	config.OverdueAlertSchedule = 24 * time.Hour
	config.ReminderAlertSchedule = 24 * time.Hour
	config.CleanupSchedule = 6 * time.Hour
	
	return NewManager(wrapper, config)
}

// AlertServiceWrapper wraps the concrete AlertService to implement our interface
type AlertServiceWrapper struct {
	alertService *services.AlertService
}

func (w *AlertServiceWrapper) GenerateOverdueAlerts() error {
	return w.alertService.GenerateOverdueAlerts()
}

func (w *AlertServiceWrapper) GenerateReminderAlerts() error {
	return w.alertService.GenerateReminderAlerts()
}

func (w *AlertServiceWrapper) CleanupResolvedAlerts() error {
	return w.alertService.CleanupResolvedAlerts()
}

// StartJobManagerWithGracefulShutdown demonstrates how to start the job manager
// with proper graceful shutdown handling
func StartJobManagerWithGracefulShutdown(manager *Manager) {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start the job manager
	if err := manager.Start(ctx); err != nil {
		log.Fatalf("Failed to start job manager: %v", err)
	}
	
	log.Println("Job manager started successfully")
	
	// In a real application, you would listen for shutdown signals here
	// For example:
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// <-c
	
	// Stop the job manager
	if err := manager.Stop(); err != nil {
		log.Printf("Error stopping job manager: %v", err)
	} else {
		log.Println("Job manager stopped successfully")
	}
}

// CustomJobExample shows how to add custom jobs to the manager
func CustomJobExample(manager *Manager) {
	// Add a custom job that runs every hour
	manager.AddCustomJob("custom-maintenance", "Custom maintenance task", 1*time.Hour, func() error {
		log.Println("Running custom maintenance task")
		
		// Your custom logic here
		// For example: database cleanup, log rotation, etc.
		
		return nil
	})
	
	// Add a job that runs every 30 minutes
	manager.AddCustomJob("health-check", "System health check", 30*time.Minute, func() error {
		log.Println("Running system health check")
		
		// Your health check logic here
		// For example: check database connectivity, disk space, etc.
		
		return nil
	})
}

// MonitoringExample shows how to monitor job execution
func MonitoringExample(manager *Manager) {
	// Get current health status
	health := manager.GetHealthStatus()
	log.Printf("Job Manager Health: Running=%v, Total Jobs=%d, Enabled=%d, Disabled=%d", 
		health.IsRunning, health.TotalJobs, health.EnabledJobs, health.DisabledJobs)
	
	// Get job statistics
	stats := manager.GetJobStatistics()
	log.Printf("Job Statistics: Total Executions=%d, Successful=%d, Failed=%d", 
		stats.TotalExecutions, stats.SuccessfulExecutions, stats.FailedExecutions)
	
	// Print success rates for each job
	for jobName, successRate := range stats.JobSuccessRates {
		log.Printf("Job '%s' success rate: %.2f%%", jobName, successRate)
	}
	
	// Get recent job executions
	executions := manager.GetJobExecutions(10)
	log.Printf("Recent executions (%d):", len(executions))
	for _, exec := range executions {
		log.Printf("  %s: %s (%s) - %s", exec.JobName, exec.Status, exec.Duration, exec.StartTime.Format(time.RFC3339))
		if exec.Error != "" {
			log.Printf("    Error: %s", exec.Error)
		}
	}
}