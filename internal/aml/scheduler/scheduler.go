package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// JobFunction is the function signature for scheduled jobs
type JobFunction func(ctx context.Context) error

// JobStatus represents the current status of a job
type JobStatus string

const (
	// JobStatusPending means the job is scheduled but hasn't run yet
	JobStatusPending JobStatus = "pending"
	// JobStatusRunning means the job is currently running
	JobStatusRunning JobStatus = "running"
	// JobStatusComplete means the job completed successfully
	JobStatusComplete JobStatus = "complete"
	// JobStatusFailed means the job failed
	JobStatusFailed JobStatus = "failed"
)

// Job represents a scheduled job
type Job struct {
	ID           string
	Name         string
	Schedule     string
	Function     JobFunction
	Status       JobStatus
	LastRun      *time.Time
	NextRun      *time.Time
	LastError    error
	DatabaseName string
	cronEntryID  cron.EntryID
}

// Scheduler manages scheduled jobs
type Scheduler struct {
	cron      *cron.Cron
	jobs      map[string]*Job
	jobsMutex sync.RWMutex
}

// NewScheduler creates a new job scheduler
func NewScheduler() *Scheduler {
	cronScheduler := cron.New(cron.WithSeconds())
	return &Scheduler{
		cron: cronScheduler,
		jobs: make(map[string]*Job),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(name, schedule string, databaseName string, fn JobFunction) (string, error) {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	// Validate the schedule
	if _, err := cron.ParseStandard(schedule); err != nil {
		return "", fmt.Errorf("invalid schedule format: %w", err)
	}

	// Generate a unique ID for the job
	jobID := uuid.New().String()

	// Create the job
	job := &Job{
		ID:           jobID,
		Name:         name,
		Schedule:     schedule,
		Function:     fn,
		Status:       JobStatusPending,
		DatabaseName: databaseName,
	}

	// Wrap the job function to update status
	wrappedFn := func() {
		// Update job status to running
		s.jobsMutex.Lock()
		job.Status = JobStatusRunning
		now := time.Now()
		job.LastRun = &now
		s.jobsMutex.Unlock()

		// Execute the job function
		ctx := context.Background()
		err := fn(ctx)

		// Update job status based on result
		s.jobsMutex.Lock()
		if err != nil {
			job.Status = JobStatusFailed
			job.LastError = err
		} else {
			job.Status = JobStatusComplete
			job.LastError = nil
		}
		s.jobsMutex.Unlock()
	}

	// Schedule the job
	entryID, err := s.cron.AddFunc(schedule, wrappedFn)
	if err != nil {
		return "", fmt.Errorf("failed to schedule job: %w", err)
	}

	// Update the job with entry ID
	job.cronEntryID = entryID

	// Update next run time
	entry := s.cron.Entry(entryID)
	job.NextRun = &entry.Next

	// Store the job
	s.jobs[jobID] = job

	return jobID, nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID string) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job with ID '%s' not found", jobID)
	}

	// Remove the job from the cron scheduler
	s.cron.Remove(job.cronEntryID)

	// Remove the job from our map
	delete(s.jobs, jobID)

	return nil
}

// GetJob returns a job by ID
func (s *Scheduler) GetJob(jobID string) (*Job, error) {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job with ID '%s' not found", jobID)
	}

	// Update next run time
	entry := s.cron.Entry(job.cronEntryID)
	nextRun := entry.Next
	job.NextRun = &nextRun

	return job, nil
}

// ListJobs returns all scheduled jobs
func (s *Scheduler) ListJobs() []*Job {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		// Update next run time
		entry := s.cron.Entry(job.cronEntryID)
		nextRun := entry.Next
		job.NextRun = &nextRun

		jobs = append(jobs, job)
	}

	return jobs
}

// RunJobNow runs a job immediately
func (s *Scheduler) RunJobNow(jobID string) error {
	s.jobsMutex.RLock()
	job, exists := s.jobs[jobID]
	if !exists {
		s.jobsMutex.RUnlock()
		return fmt.Errorf("job with ID '%s' not found", jobID)
	}

	// Get a reference to the job function
	fn := job.Function
	s.jobsMutex.RUnlock()

	// Update job status to running
	s.jobsMutex.Lock()
	job.Status = JobStatusRunning
	now := time.Now()
	job.LastRun = &now
	s.jobsMutex.Unlock()

	// Execute the job function in a goroutine
	go func() {
		ctx := context.Background()
		err := fn(ctx)

		// Update job status based on result
		s.jobsMutex.Lock()
		defer s.jobsMutex.Unlock()

		if err != nil {
			job.Status = JobStatusFailed
			job.LastError = err
		} else {
			job.Status = JobStatusComplete
			job.LastError = nil
		}
	}()

	return nil
}