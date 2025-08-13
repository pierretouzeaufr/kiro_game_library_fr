package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// JobStatus represents the status of a job execution
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// JobExecution represents a single job execution
type JobExecution struct {
	ID        string    `json:"id"`
	JobName   string    `json:"job_name"`
	Status    JobStatus `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration"`
}

// Job represents a scheduled job
type Job struct {
	Name        string
	Description string
	Schedule    time.Duration
	Handler     func() error
	LastRun     time.Time
	NextRun     time.Time
	Enabled     bool
}

// Scheduler manages background jobs
type Scheduler struct {
	jobs        map[string]*Job
	executions  []JobExecution
	alertService AlertService
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	logger      *log.Logger
}

// NewScheduler creates a new job scheduler
func NewScheduler(alertService AlertService, logger *log.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	if logger == nil {
		logger = log.New(log.Writer(), "[SCHEDULER] ", log.LstdFlags)
	}
	
	scheduler := &Scheduler{
		jobs:         make(map[string]*Job),
		executions:   make([]JobExecution, 0),
		alertService: alertService,
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
	}
	
	// Register default alert jobs
	scheduler.registerAlertJobs()
	
	return scheduler
}

// registerAlertJobs registers the default alert generation jobs
func (s *Scheduler) registerAlertJobs() {
	// Daily overdue alert generation job
	s.AddJob("overdue-alerts", "Generate alerts for overdue items", 24*time.Hour, func() error {
		s.logger.Println("Starting overdue alert generation")
		err := s.alertService.GenerateOverdueAlerts()
		if err != nil {
			s.logger.Printf("Failed to generate overdue alerts: %v", err)
			return fmt.Errorf("overdue alert generation failed: %w", err)
		}
		s.logger.Println("Overdue alert generation completed successfully")
		return nil
	})
	
	// Daily reminder alert generation job
	s.AddJob("reminder-alerts", "Generate reminder alerts for items due soon", 24*time.Hour, func() error {
		s.logger.Println("Starting reminder alert generation")
		err := s.alertService.GenerateReminderAlerts()
		if err != nil {
			s.logger.Printf("Failed to generate reminder alerts: %v", err)
			return fmt.Errorf("reminder alert generation failed: %w", err)
		}
		s.logger.Println("Reminder alert generation completed successfully")
		return nil
	})
	
	// Cleanup resolved alerts job (runs every 6 hours)
	s.AddJob("cleanup-alerts", "Clean up alerts for returned items", 6*time.Hour, func() error {
		s.logger.Println("Starting alert cleanup")
		err := s.alertService.CleanupResolvedAlerts()
		if err != nil {
			s.logger.Printf("Failed to cleanup resolved alerts: %v", err)
			return fmt.Errorf("alert cleanup failed: %w", err)
		}
		s.logger.Println("Alert cleanup completed successfully")
		return nil
	})
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(name, description string, schedule time.Duration, handler func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	job := &Job{
		Name:        name,
		Description: description,
		Schedule:    schedule,
		Handler:     handler,
		NextRun:     time.Now().Add(schedule),
		Enabled:     true,
	}
	
	s.jobs[name] = job
	s.logger.Printf("Added job '%s' with schedule %v", name, schedule)
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.jobs, name)
	s.logger.Printf("Removed job '%s'", name)
}

// EnableJob enables a job
func (s *Scheduler) EnableJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}
	
	job.Enabled = true
	s.logger.Printf("Enabled job '%s'", name)
	return nil
}

// DisableJob disables a job
func (s *Scheduler) DisableJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}
	
	job.Enabled = false
	s.logger.Printf("Disabled job '%s'", name)
	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.logger.Println("Starting job scheduler")
	
	s.wg.Add(1)
	go s.run()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Println("Stopping job scheduler")
	s.cancel()
	s.wg.Wait()
	s.logger.Println("Job scheduler stopped")
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkAndRunJobs()
		}
	}
}

// checkAndRunJobs checks if any jobs need to run and executes them
func (s *Scheduler) checkAndRunJobs() {
	s.mu.RLock()
	jobsToRun := make([]*Job, 0)
	now := time.Now()
	
	for _, job := range s.jobs {
		if job.Enabled && now.After(job.NextRun) {
			jobsToRun = append(jobsToRun, job)
		}
	}
	s.mu.RUnlock()
	
	// Run jobs outside of the lock to avoid blocking
	for _, job := range jobsToRun {
		s.runJob(job)
	}
}

// runJob executes a single job
func (s *Scheduler) runJob(job *Job) {
	executionID := fmt.Sprintf("%s-%d", job.Name, time.Now().Unix())
	
	execution := JobExecution{
		ID:        executionID,
		JobName:   job.Name,
		Status:    JobStatusRunning,
		StartTime: time.Now(),
	}
	
	s.mu.Lock()
	s.executions = append(s.executions, execution)
	executionIndex := len(s.executions) - 1
	s.mu.Unlock()
	
	s.logger.Printf("Starting job execution: %s", executionID)
	
	// Run the job
	err := job.Handler()
	
	// Update execution record
	s.mu.Lock()
	s.executions[executionIndex].EndTime = time.Now()
	s.executions[executionIndex].Duration = s.executions[executionIndex].EndTime.Sub(s.executions[executionIndex].StartTime).String()
	
	if err != nil {
		s.executions[executionIndex].Status = JobStatusFailed
		s.executions[executionIndex].Error = err.Error()
		s.logger.Printf("Job execution failed: %s - %v", executionID, err)
	} else {
		s.executions[executionIndex].Status = JobStatusCompleted
		s.logger.Printf("Job execution completed: %s", executionID)
	}
	
	// Update job's next run time
	job.LastRun = time.Now()
	job.NextRun = job.LastRun.Add(job.Schedule)
	
	// Keep only the last 100 executions to prevent memory growth
	if len(s.executions) > 100 {
		s.executions = s.executions[len(s.executions)-100:]
	}
	
	s.mu.Unlock()
}

// RunJobNow executes a job immediately, regardless of schedule
func (s *Scheduler) RunJobNow(name string) error {
	s.mu.RLock()
	job, exists := s.jobs[name]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}
	
	if !job.Enabled {
		return fmt.Errorf("job '%s' is disabled", name)
	}
	
	go s.runJob(job)
	return nil
}

// GetJobs returns all registered jobs
func (s *Scheduler) GetJobs() map[string]*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	jobs := make(map[string]*Job)
	for name, job := range s.jobs {
		// Create a copy to avoid race conditions
		jobCopy := *job
		jobs[name] = &jobCopy
	}
	
	return jobs
}

// GetJobExecutions returns recent job executions
func (s *Scheduler) GetJobExecutions(limit int) []JobExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if limit <= 0 || limit > len(s.executions) {
		limit = len(s.executions)
	}
	
	// Return the most recent executions
	start := len(s.executions) - limit
	if start < 0 {
		start = 0
	}
	
	executions := make([]JobExecution, limit)
	copy(executions, s.executions[start:])
	
	// Reverse to show most recent first
	for i, j := 0, len(executions)-1; i < j; i, j = i+1, j-1 {
		executions[i], executions[j] = executions[j], executions[i]
	}
	
	return executions
}

// GetJobStatus returns the status of a specific job
func (s *Scheduler) GetJobStatus(name string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	job, exists := s.jobs[name]
	if !exists {
		return nil, fmt.Errorf("job '%s' not found", name)
	}
	
	// Return a copy to avoid race conditions
	jobCopy := *job
	return &jobCopy, nil
}