# Background Job System

This package provides a comprehensive background job system for the Board Game Library application, specifically designed to handle alert generation and cleanup tasks.

## Features

- **Scheduled Job Execution**: Run jobs on configurable schedules
- **Job Management**: Add, remove, enable, disable jobs dynamically
- **Error Handling**: Comprehensive error handling with logging
- **Execution Tracking**: Track job execution history and statistics
- **Graceful Shutdown**: Proper cleanup when stopping the scheduler
- **Thread Safety**: Safe for concurrent access
- **Monitoring**: Health status and performance metrics

## Core Components

### Scheduler
The `Scheduler` is the core component that manages job execution:
- Runs jobs on their configured schedules
- Tracks execution history
- Handles errors and logging
- Provides thread-safe operations

### Manager
The `Manager` provides a high-level interface for job management:
- Configurable job setup
- Lifecycle management (start/stop)
- Health monitoring
- Statistics collection

### Job
A `Job` represents a scheduled task:
- Name and description
- Schedule (time.Duration)
- Handler function
- Enable/disable state
- Last run and next run times

## Default Alert Jobs

The system comes with three pre-configured alert jobs:

1. **Overdue Alerts** (`overdue-alerts`)
   - Generates alerts for overdue borrowed items
   - Default schedule: 24 hours
   - Calls `AlertService.GenerateOverdueAlerts()`

2. **Reminder Alerts** (`reminder-alerts`)
   - Generates reminder alerts for items due within 2 days
   - Default schedule: 24 hours
   - Calls `AlertService.GenerateReminderAlerts()`

3. **Alert Cleanup** (`cleanup-alerts`)
   - Cleans up alerts for returned items
   - Default schedule: 6 hours
   - Calls `AlertService.CleanupResolvedAlerts()`

## Usage

### Basic Setup

```go
package main

import (
    "board-game-library/internal/jobs"
    "board-game-library/internal/services"
    "context"
    "log"
)

func main() {
    // Create your alert service
    alertService := services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)
    
    // Wrap it to implement the jobs.AlertService interface
    wrapper := &jobs.AlertServiceWrapper{AlertService: alertService}
    
    // Create job manager with default configuration
    manager := jobs.NewManager(wrapper, nil)
    
    // Start the job manager
    ctx := context.Background()
    if err := manager.Start(ctx); err != nil {
        log.Fatalf("Failed to start job manager: %v", err)
    }
    defer manager.Stop()
    
    // Jobs will now run automatically in the background
}
```

### Custom Configuration

```go
config := &jobs.Config{
    EnableOverdueAlerts:   true,
    EnableReminderAlerts:  true,
    EnableAlertCleanup:    true,
    OverdueAlertSchedule:  12 * time.Hour, // Run twice daily
    ReminderAlertSchedule: 24 * time.Hour, // Run daily
    CleanupSchedule:       3 * time.Hour,  // Run every 3 hours
    Logger:                log.New(os.Stdout, "[JOBS] ", log.LstdFlags),
}

manager := jobs.NewManager(alertService, config)
```

### Adding Custom Jobs

```go
// Add a custom maintenance job
manager.AddCustomJob("maintenance", "Database maintenance", 6*time.Hour, func() error {
    log.Println("Running database maintenance")
    // Your maintenance logic here
    return nil
})

// Add a health check job
manager.AddCustomJob("health-check", "System health check", 30*time.Minute, func() error {
    // Your health check logic here
    return checkSystemHealth()
})
```

### Manual Job Execution

```go
// Run a specific job immediately
err := manager.RunJobNow("overdue-alerts")
if err != nil {
    log.Printf("Failed to run job: %v", err)
}

// Run all alert generation jobs
err = manager.GenerateAllAlerts()
if err != nil {
    log.Printf("Failed to generate alerts: %v", err)
}

// Run alert cleanup
err = manager.CleanupAlerts()
if err != nil {
    log.Printf("Failed to cleanup alerts: %v", err)
}
```

### Job Management

```go
// Enable/disable jobs
manager.EnableJob("overdue-alerts")
manager.DisableJob("reminder-alerts")

// Remove a job
manager.RemoveJob("custom-job")

// Get job status
job, err := manager.GetJobStatus("overdue-alerts")
if err == nil {
    log.Printf("Job: %s, Enabled: %v, Last Run: %v, Next Run: %v", 
        job.Name, job.Enabled, job.LastRun, job.NextRun)
}
```

### Monitoring and Health Checks

```go
// Get health status
health := manager.GetHealthStatus()
log.Printf("Running: %v, Total Jobs: %d, Enabled: %d", 
    health.IsRunning, health.TotalJobs, health.EnabledJobs)

// Get execution statistics
stats := manager.GetJobStatistics()
log.Printf("Total Executions: %d, Success Rate: %.2f%%", 
    stats.TotalExecutions, 
    float64(stats.SuccessfulExecutions)/float64(stats.TotalExecutions)*100)

// Get recent executions
executions := manager.GetJobExecutions(10)
for _, exec := range executions {
    log.Printf("%s: %s (%s)", exec.JobName, exec.Status, exec.Duration)
}
```

### Graceful Shutdown

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func main() {
    manager := jobs.NewManager(alertService, nil)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Start job manager
    manager.Start(ctx)
    
    // Wait for shutdown signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    log.Println("Shutting down...")
    
    // Stop job manager gracefully
    if err := manager.Stop(); err != nil {
        log.Printf("Error stopping job manager: %v", err)
    }
}
```

## Configuration Options

### Config Struct

```go
type Config struct {
    EnableOverdueAlerts   bool          // Enable overdue alert generation
    EnableReminderAlerts  bool          // Enable reminder alert generation
    EnableAlertCleanup    bool          // Enable alert cleanup
    OverdueAlertSchedule  time.Duration // Schedule for overdue alerts
    ReminderAlertSchedule time.Duration // Schedule for reminder alerts
    CleanupSchedule       time.Duration // Schedule for cleanup
    Logger                *log.Logger   // Logger for job operations
}
```

### Default Values

- `EnableOverdueAlerts`: `true`
- `EnableReminderAlerts`: `true`
- `EnableAlertCleanup`: `true`
- `OverdueAlertSchedule`: `24 * time.Hour`
- `ReminderAlertSchedule`: `24 * time.Hour`
- `CleanupSchedule`: `6 * time.Hour`

## Error Handling

The job system includes comprehensive error handling:

- Job execution errors are logged and tracked
- Failed jobs don't affect other jobs
- Execution history includes error details
- Jobs continue to be scheduled even after failures

## Thread Safety

All operations are thread-safe:
- Multiple goroutines can safely call manager methods
- Job execution is handled in separate goroutines
- Internal state is protected with mutexes

## Performance Considerations

- Execution history is limited to 100 entries to prevent memory growth
- Jobs run in separate goroutines to avoid blocking
- Scheduler checks for jobs to run every minute
- Minimal overhead when no jobs need to run

## Testing

The package includes comprehensive tests:
- Unit tests for all components
- Mock implementations for testing
- Concurrent access testing
- Error handling testing

Run tests with:
```bash
go test ./internal/jobs -v
```

## Integration with Main Application

See `example_integration.go` for complete integration examples showing how to:
- Set up the job manager in your main application
- Handle graceful shutdown
- Add custom jobs
- Monitor job execution
- Handle errors and logging