package jobs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewScheduler(t *testing.T) {
	mockAlertService := &MockAlertService{}
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	
	scheduler := NewScheduler(mockAlertService, logger)
	
	assert.NotNil(t, scheduler)
	assert.Equal(t, mockAlertService, scheduler.alertService)
	assert.Equal(t, logger, scheduler.logger)
	assert.NotNil(t, scheduler.jobs)
	assert.NotNil(t, scheduler.executions)
	assert.NotNil(t, scheduler.ctx)
	assert.NotNil(t, scheduler.cancel)
	
	// Check that default jobs are registered
	jobs := scheduler.GetJobs()
	assert.Contains(t, jobs, "overdue-alerts")
	assert.Contains(t, jobs, "reminder-alerts")
	assert.Contains(t, jobs, "cleanup-alerts")
	
	// Verify job properties
	overdueJob := jobs["overdue-alerts"]
	assert.Equal(t, "Generate alerts for overdue items", overdueJob.Description)
	assert.Equal(t, 24*time.Hour, overdueJob.Schedule)
	assert.True(t, overdueJob.Enabled)
	
	reminderJob := jobs["reminder-alerts"]
	assert.Equal(t, "Generate reminder alerts for items due soon", reminderJob.Description)
	assert.Equal(t, 24*time.Hour, reminderJob.Schedule)
	assert.True(t, reminderJob.Enabled)
	
	cleanupJob := jobs["cleanup-alerts"]
	assert.Equal(t, "Clean up alerts for returned items", cleanupJob.Description)
	assert.Equal(t, 6*time.Hour, cleanupJob.Schedule)
	assert.True(t, cleanupJob.Enabled)
}

func TestScheduler_AddJob(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	jobExecuted := false
	handler := func() error {
		jobExecuted = true
		return nil
	}
	
	scheduler.AddJob("test-job", "Test job description", 1*time.Hour, handler)
	
	jobs := scheduler.GetJobs()
	assert.Contains(t, jobs, "test-job")
	
	testJob := jobs["test-job"]
	assert.Equal(t, "test-job", testJob.Name)
	assert.Equal(t, "Test job description", testJob.Description)
	assert.Equal(t, 1*time.Hour, testJob.Schedule)
	assert.True(t, testJob.Enabled)
	assert.NotNil(t, testJob.Handler)
	
	// Test that handler works
	err := testJob.Handler()
	assert.NoError(t, err)
	assert.True(t, jobExecuted)
}

func TestScheduler_RemoveJob(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Add a test job
	scheduler.AddJob("test-job", "Test job", 1*time.Hour, func() error { return nil })
	
	// Verify job exists
	jobs := scheduler.GetJobs()
	assert.Contains(t, jobs, "test-job")
	
	// Remove job
	scheduler.RemoveJob("test-job")
	
	// Verify job is removed
	jobs = scheduler.GetJobs()
	assert.NotContains(t, jobs, "test-job")
}

func TestScheduler_EnableDisableJob(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Test disabling existing job
	err := scheduler.DisableJob("overdue-alerts")
	assert.NoError(t, err)
	
	jobs := scheduler.GetJobs()
	assert.False(t, jobs["overdue-alerts"].Enabled)
	
	// Test enabling job
	err = scheduler.EnableJob("overdue-alerts")
	assert.NoError(t, err)
	
	jobs = scheduler.GetJobs()
	assert.True(t, jobs["overdue-alerts"].Enabled)
	
	// Test enabling non-existent job
	err = scheduler.EnableJob("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job 'non-existent' not found")
	
	// Test disabling non-existent job
	err = scheduler.DisableJob("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job 'non-existent' not found")
}

func TestScheduler_RunJobNow(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	jobExecuted := false
	scheduler.AddJob("test-job", "Test job", 1*time.Hour, func() error {
		jobExecuted = true
		return nil
	})
	
	// Test running existing job
	err := scheduler.RunJobNow("test-job")
	assert.NoError(t, err)
	
	// Give some time for the goroutine to execute
	time.Sleep(100 * time.Millisecond)
	assert.True(t, jobExecuted)
	
	// Test running non-existent job
	err = scheduler.RunJobNow("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job 'non-existent' not found")
	
	// Test running disabled job
	scheduler.DisableJob("test-job")
	err = scheduler.RunJobNow("test-job")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job 'test-job' is disabled")
}

func TestScheduler_GetJobStatus(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Test getting status of existing job
	job, err := scheduler.GetJobStatus("overdue-alerts")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "overdue-alerts", job.Name)
	assert.Equal(t, "Generate alerts for overdue items", job.Description)
	
	// Test getting status of non-existent job
	job, err = scheduler.GetJobStatus("non-existent")
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "job 'non-existent' not found")
}

func TestScheduler_JobExecution(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	executionCount := 0
	scheduler.AddJob("test-job", "Test job", 1*time.Hour, func() error {
		executionCount++
		return nil
	})
	
	// Run job manually to test execution tracking
	err := scheduler.RunJobNow("test-job")
	assert.NoError(t, err)
	
	// Wait for job to complete
	time.Sleep(100 * time.Millisecond)
	
	// Check that job executed
	assert.Greater(t, executionCount, 0)
	
	// Check execution history
	executions := scheduler.GetJobExecutions(10)
	assert.Greater(t, len(executions), 0)
	
	// Verify execution details
	execution := executions[0]
	assert.Equal(t, "test-job", execution.JobName)
	assert.Equal(t, JobStatusCompleted, execution.Status)
	assert.NotEmpty(t, execution.ID)
	assert.NotZero(t, execution.StartTime)
	assert.NotZero(t, execution.EndTime)
	assert.NotEmpty(t, execution.Duration)
	assert.Empty(t, execution.Error)
}

func TestScheduler_JobExecutionWithError(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	testError := errors.New("test error")
	scheduler.AddJob("failing-job", "Failing job", 1*time.Hour, func() error {
		return testError
	})
	
	// Run job manually to test error handling
	err := scheduler.RunJobNow("failing-job")
	assert.NoError(t, err) // RunJobNow doesn't return job execution errors
	
	// Wait for job to complete
	time.Sleep(100 * time.Millisecond)
	
	// Check execution history
	executions := scheduler.GetJobExecutions(10)
	assert.Greater(t, len(executions), 0)
	
	// Find the failing job execution
	var failedExecution *JobExecution
	for _, exec := range executions {
		if exec.JobName == "failing-job" {
			failedExecution = &exec
			break
		}
	}
	
	assert.NotNil(t, failedExecution)
	assert.Equal(t, JobStatusFailed, failedExecution.Status)
	assert.Contains(t, failedExecution.Error, "test error")
}

func TestScheduler_AlertJobsExecution(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Set up mock expectations
	mockAlertService.On("GenerateOverdueAlerts").Return(nil)
	mockAlertService.On("GenerateReminderAlerts").Return(nil)
	mockAlertService.On("CleanupResolvedAlerts").Return(nil)
	
	// Test overdue alerts job
	err := scheduler.RunJobNow("overdue-alerts")
	assert.NoError(t, err)
	
	// Test reminder alerts job
	err = scheduler.RunJobNow("reminder-alerts")
	assert.NoError(t, err)
	
	// Test cleanup alerts job
	err = scheduler.RunJobNow("cleanup-alerts")
	assert.NoError(t, err)
	
	// Give time for goroutines to complete
	time.Sleep(100 * time.Millisecond)
	
	// Verify mock expectations
	mockAlertService.AssertExpectations(t)
}

func TestScheduler_AlertJobsExecutionWithErrors(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Set up mock expectations with errors
	mockAlertService.On("GenerateOverdueAlerts").Return(errors.New("overdue error"))
	mockAlertService.On("GenerateReminderAlerts").Return(errors.New("reminder error"))
	mockAlertService.On("CleanupResolvedAlerts").Return(errors.New("cleanup error"))
	
	// Test jobs with errors
	err := scheduler.RunJobNow("overdue-alerts")
	assert.NoError(t, err) // RunJobNow doesn't return job execution errors
	
	err = scheduler.RunJobNow("reminder-alerts")
	assert.NoError(t, err)
	
	err = scheduler.RunJobNow("cleanup-alerts")
	assert.NoError(t, err)
	
	// Give time for goroutines to complete
	time.Sleep(100 * time.Millisecond)
	
	// Check execution history for errors
	executions := scheduler.GetJobExecutions(10)
	assert.Greater(t, len(executions), 0)
	
	// Verify that jobs failed
	failedJobs := make(map[string]bool)
	for _, exec := range executions {
		if exec.Status == JobStatusFailed {
			failedJobs[exec.JobName] = true
		}
	}
	
	assert.True(t, failedJobs["overdue-alerts"])
	assert.True(t, failedJobs["reminder-alerts"])
	assert.True(t, failedJobs["cleanup-alerts"])
	
	// Verify mock expectations
	mockAlertService.AssertExpectations(t)
}

func TestScheduler_GetJobExecutions(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Add multiple jobs and run them
	for i := 0; i < 5; i++ {
		jobName := fmt.Sprintf("test-job-%d", i)
		scheduler.AddJob(jobName, "Test job", 1*time.Hour, func() error {
			return nil
		})
		scheduler.RunJobNow(jobName)
	}
	
	// Give time for jobs to complete
	time.Sleep(100 * time.Millisecond)
	
	// Test getting limited executions
	executions := scheduler.GetJobExecutions(3)
	assert.Len(t, executions, 3)
	
	// Test getting all executions
	allExecutions := scheduler.GetJobExecutions(0)
	assert.GreaterOrEqual(t, len(allExecutions), 5)
	
	// Test getting more executions than available
	moreExecutions := scheduler.GetJobExecutions(100)
	assert.Equal(t, len(allExecutions), len(moreExecutions))
}

func TestScheduler_StartStop(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Test starting and stopping scheduler
	scheduler.Start()
	
	// Verify scheduler is running by checking context
	select {
	case <-scheduler.ctx.Done():
		t.Fatal("Scheduler context should not be done after start")
	default:
		// Context is not done, scheduler is running
	}
	
	// Stop scheduler
	scheduler.Stop()
	
	// Verify scheduler is stopped
	select {
	case <-scheduler.ctx.Done():
		// Context is done, scheduler is stopped
	case <-time.After(1 * time.Second):
		t.Fatal("Scheduler should be stopped")
	}
}

func TestScheduler_ConcurrentAccess(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Test concurrent job additions and removals
	done := make(chan bool, 10)
	
	// Add jobs concurrently
	for i := 0; i < 5; i++ {
		go func(id int) {
			jobName := fmt.Sprintf("concurrent-job-%d", id)
			scheduler.AddJob(jobName, "Concurrent job", 1*time.Hour, func() error {
				return nil
			})
			done <- true
		}(i)
	}
	
	// Remove jobs concurrently
	for i := 0; i < 5; i++ {
		go func(id int) {
			jobName := fmt.Sprintf("concurrent-job-%d", id)
			time.Sleep(10 * time.Millisecond) // Small delay to ensure job is added first
			scheduler.RemoveJob(jobName)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify no race conditions occurred (test passes if no panic)
	jobs := scheduler.GetJobs()
	assert.NotNil(t, jobs)
}

func TestScheduler_ExecutionHistoryLimit(t *testing.T) {
	mockAlertService := &MockAlertService{}
	scheduler := NewScheduler(mockAlertService, nil)
	
	// Add a job that runs quickly
	scheduler.AddJob("history-test", "History test job", 1*time.Millisecond, func() error {
		return nil
	})
	
	// Start scheduler and let it run many executions
	scheduler.Start()
	time.Sleep(200 * time.Millisecond) // Let it run for a while
	scheduler.Stop()
	
	// Check that execution history is limited
	executions := scheduler.GetJobExecutions(0)
	assert.LessOrEqual(t, len(executions), 100, "Execution history should be limited to 100 entries")
}