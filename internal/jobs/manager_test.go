package jobs

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.True(t, config.EnableOverdueAlerts)
	assert.True(t, config.EnableReminderAlerts)
	assert.True(t, config.EnableAlertCleanup)
	assert.Equal(t, 24*time.Hour, config.OverdueAlertSchedule)
	assert.Equal(t, 24*time.Hour, config.ReminderAlertSchedule)
	assert.Equal(t, 6*time.Hour, config.CleanupSchedule)
	assert.NotNil(t, config.Logger)
}

func TestNewManager(t *testing.T) {
	mockAlertService := &MockAlertService{}
	
	// Test with default config
	manager := NewManager(mockAlertService, nil)
	assert.NotNil(t, manager)
	assert.Equal(t, mockAlertService, manager.alertService)
	assert.NotNil(t, manager.scheduler)
	assert.NotNil(t, manager.logger)
	assert.False(t, manager.started)
	
	// Verify default jobs are configured
	jobs := manager.GetJobs()
	assert.Contains(t, jobs, "overdue-alerts")
	assert.Contains(t, jobs, "reminder-alerts")
	assert.Contains(t, jobs, "cleanup-alerts")
}

func TestNewManagerWithCustomConfig(t *testing.T) {
	mockAlertService := &MockAlertService{}
	
	config := &Config{
		EnableOverdueAlerts:   true,
		EnableReminderAlerts:  false,
		EnableAlertCleanup:    true,
		OverdueAlertSchedule:  12 * time.Hour,
		ReminderAlertSchedule: 6 * time.Hour,
		CleanupSchedule:       3 * time.Hour,
		Logger:                log.New(os.Stdout, "[TEST] ", log.LstdFlags),
	}
	
	manager := NewManager(mockAlertService, config)
	assert.NotNil(t, manager)
	
	jobs := manager.GetJobs()
	assert.Contains(t, jobs, "overdue-alerts")
	assert.NotContains(t, jobs, "reminder-alerts") // Disabled in config
	assert.Contains(t, jobs, "cleanup-alerts")
	
	// Verify custom schedules
	overdueJob := jobs["overdue-alerts"]
	assert.Equal(t, 12*time.Hour, overdueJob.Schedule)
	
	cleanupJob := jobs["cleanup-alerts"]
	assert.Equal(t, 3*time.Hour, cleanupJob.Schedule)
}

func TestManagerStartStop(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Test initial state
	assert.False(t, manager.IsStarted())
	
	// Test starting
	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, manager.IsStarted())
	
	// Test starting again (should fail)
	err = manager.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")
	
	// Test stopping
	err = manager.Stop()
	assert.NoError(t, err)
	assert.False(t, manager.IsStarted())
	
	// Test stopping again (should fail)
	err = manager.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestManagerContextCancellation(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start manager
	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, manager.IsStarted())
	
	// Cancel context
	cancel()
	
	// Give some time for the context cancellation to be processed
	time.Sleep(100 * time.Millisecond)
	
	// Manager should be stopped
	assert.False(t, manager.IsStarted())
}

func TestManagerJobOperations(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	// Test adding custom job
	jobExecuted := false
	manager.AddCustomJob("test-job", "Test job", 1*time.Hour, func() error {
		jobExecuted = true
		return nil
	})
	
	jobs := manager.GetJobs()
	assert.Contains(t, jobs, "test-job")
	
	// Test running job now
	err := manager.RunJobNow("test-job")
	assert.NoError(t, err)
	
	// Give time for job to execute
	time.Sleep(100 * time.Millisecond)
	assert.True(t, jobExecuted)
	
	// Test getting job status
	job, err := manager.GetJobStatus("test-job")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-job", job.Name)
	
	// Test disabling job
	err = manager.DisableJob("test-job")
	assert.NoError(t, err)
	
	job, err = manager.GetJobStatus("test-job")
	assert.NoError(t, err)
	assert.False(t, job.Enabled)
	
	// Test enabling job
	err = manager.EnableJob("test-job")
	assert.NoError(t, err)
	
	job, err = manager.GetJobStatus("test-job")
	assert.NoError(t, err)
	assert.True(t, job.Enabled)
	
	// Test removing job
	manager.RemoveJob("test-job")
	
	jobs = manager.GetJobs()
	assert.NotContains(t, jobs, "test-job")
}

func TestManagerGetHealthStatus(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	// Test health status when not started
	health := manager.GetHealthStatus()
	assert.False(t, health.IsRunning)
	assert.Equal(t, 3, health.TotalJobs) // Default jobs: overdue, reminder, cleanup
	assert.Equal(t, 3, health.EnabledJobs)
	assert.Equal(t, 0, health.DisabledJobs)
	assert.Nil(t, health.LastExecution)
	
	// Start manager and run a job
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := manager.Start(ctx)
	assert.NoError(t, err)
	
	// Add and run a test job
	manager.AddCustomJob("health-test", "Health test job", 1*time.Hour, func() error {
		return nil
	})
	
	err = manager.RunJobNow("health-test")
	assert.NoError(t, err)
	
	// Give time for job to execute
	time.Sleep(100 * time.Millisecond)
	
	// Test health status when started
	health = manager.GetHealthStatus()
	assert.True(t, health.IsRunning)
	assert.Equal(t, 4, health.TotalJobs) // 3 default + 1 custom
	assert.Equal(t, 4, health.EnabledJobs)
	assert.Equal(t, 0, health.DisabledJobs)
	assert.NotNil(t, health.LastExecution)
	assert.Equal(t, "health-test", health.LastExecution.JobName)
	
	// Disable a job and check health
	err = manager.DisableJob("health-test")
	assert.NoError(t, err)
	
	health = manager.GetHealthStatus()
	assert.Equal(t, 3, health.EnabledJobs)
	assert.Equal(t, 1, health.DisabledJobs)
}

func TestManagerGenerateAllAlerts(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	// Set up mock expectations
	mockAlertService.On("GenerateOverdueAlerts").Return(nil)
	mockAlertService.On("GenerateReminderAlerts").Return(nil)
	
	// Test generating all alerts
	err := manager.GenerateAllAlerts()
	assert.NoError(t, err)
	
	// Give time for jobs to execute
	time.Sleep(100 * time.Millisecond)
	
	// Verify mock expectations
	mockAlertService.AssertExpectations(t)
}

func TestManagerCleanupAlerts(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	// Set up mock expectations
	mockAlertService.On("CleanupResolvedAlerts").Return(nil)
	
	// Test cleanup alerts
	err := manager.CleanupAlerts()
	assert.NoError(t, err)
	
	// Give time for job to execute
	time.Sleep(100 * time.Millisecond)
	
	// Verify mock expectations
	mockAlertService.AssertExpectations(t)
}

func TestManagerGetJobStatistics(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	// Add test jobs with different outcomes
	manager.AddCustomJob("success-job", "Always succeeds", 1*time.Hour, func() error {
		return nil
	})
	
	manager.AddCustomJob("fail-job", "Always fails", 1*time.Hour, func() error {
		return assert.AnError
	})
	
	// Run jobs multiple times
	for i := 0; i < 3; i++ {
		manager.RunJobNow("success-job")
		manager.RunJobNow("fail-job")
	}
	
	// Give time for jobs to execute
	time.Sleep(200 * time.Millisecond)
	
	// Get statistics
	stats := manager.GetJobStatistics()
	
	assert.Greater(t, stats.TotalExecutions, 0)
	assert.Greater(t, stats.SuccessfulExecutions, 0)
	assert.Greater(t, stats.FailedExecutions, 0)
	assert.Contains(t, stats.JobExecutionCounts, "success-job")
	assert.Contains(t, stats.JobExecutionCounts, "fail-job")
	assert.Contains(t, stats.JobSuccessRates, "success-job")
	assert.Contains(t, stats.JobSuccessRates, "fail-job")
	
	// Success job should have 100% success rate
	assert.Equal(t, 100.0, stats.JobSuccessRates["success-job"])
	
	// Fail job should have 0% success rate
	assert.Equal(t, 0.0, stats.JobSuccessRates["fail-job"])
}

func TestManagerGetJobExecutions(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	// Add and run test jobs
	manager.AddCustomJob("exec-test-1", "Execution test 1", 1*time.Hour, func() error {
		return nil
	})
	
	manager.AddCustomJob("exec-test-2", "Execution test 2", 1*time.Hour, func() error {
		return nil
	})
	
	// Run jobs
	manager.RunJobNow("exec-test-1")
	manager.RunJobNow("exec-test-2")
	
	// Give time for jobs to execute
	time.Sleep(100 * time.Millisecond)
	
	// Test getting executions
	executions := manager.GetJobExecutions(10)
	assert.Greater(t, len(executions), 0)
	
	// Verify execution details
	for _, exec := range executions {
		assert.NotEmpty(t, exec.ID)
		assert.NotEmpty(t, exec.JobName)
		assert.NotZero(t, exec.StartTime)
		assert.NotZero(t, exec.EndTime)
		assert.NotEmpty(t, exec.Duration)
	}
}

func TestManagerConfigureJobsDisabled(t *testing.T) {
	mockAlertService := &MockAlertService{}
	
	// Create config with all jobs disabled
	config := &Config{
		EnableOverdueAlerts:  false,
		EnableReminderAlerts: false,
		EnableAlertCleanup:   false,
		Logger:               log.New(os.Stdout, "[TEST] ", log.LstdFlags),
	}
	
	manager := NewManager(mockAlertService, config)
	
	// Verify no default jobs are configured
	jobs := manager.GetJobs()
	assert.NotContains(t, jobs, "overdue-alerts")
	assert.NotContains(t, jobs, "reminder-alerts")
	assert.NotContains(t, jobs, "cleanup-alerts")
	assert.Len(t, jobs, 0)
}

func TestManagerJobOperationsErrors(t *testing.T) {
	mockAlertService := &MockAlertService{}
	manager := NewManager(mockAlertService, nil)
	
	// Test running non-existent job
	err := manager.RunJobNow("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test getting status of non-existent job
	job, err := manager.GetJobStatus("non-existent")
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "not found")
	
	// Test enabling non-existent job
	err = manager.EnableJob("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test disabling non-existent job
	err = manager.DisableJob("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}