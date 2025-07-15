package service

import (
	"errors"
	"sync"
	"time"

	"todo-agent-backend/internal/logger"
	"todo-agent-backend/internal/models"

	"go.uber.org/zap"
)

var (
	ErrJobNotFound = errors.New("job not found")
)

// JobService manages job lifecycle
type JobService struct {
	jobs   map[string]*models.Job
	mutex  sync.RWMutex
	logger *logger.Logger
}

// NewJobService creates a new job service
func NewJobService(logger *logger.Logger) *JobService {
	return &JobService{
		jobs:   make(map[string]*models.Job),
		logger: logger,
	}
}

// SubmitJob submits a new job
func (js *JobService) SubmitJob(job *models.Job) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	js.jobs[job.ID] = job
	js.logger.Info("Job submitted", zap.String("job_id", job.ID))
	
	return nil
}

// GetJob retrieves a job by ID
func (js *JobService) GetJob(jobID string) (*models.Job, error) {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	job, exists := js.jobs[jobID]
	if !exists {
		return nil, ErrJobNotFound
	}

	return job, nil
}

// UpdateJob updates job status and result
func (js *JobService) UpdateJob(jobID string, status models.JobStatusEnum, result *models.ProcessingResult, errorMsg string) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	job, exists := js.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	job.Status = status
	job.Result = result
	job.Error = errorMsg
	job.UpdatedAt = time.Now()

	js.logger.Info("Job updated",
		zap.String("job_id", jobID),
		zap.String("status", string(status)),
	)

	return nil
}

// ListJobs returns all jobs for a user
func (js *JobService) ListJobs(userID string) []*models.Job {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	var userJobs []*models.Job
	for _, job := range js.jobs {
		if job.UserID == userID {
			userJobs = append(userJobs, job)
		}
	}

	return userJobs
}

// CleanupOldJobs removes jobs older than the specified duration
func (js *JobService) CleanupOldJobs(maxAge time.Duration) {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	deleted := 0

	for jobID, job := range js.jobs {
		if job.CreatedAt.Before(cutoff) {
			delete(js.jobs, jobID)
			deleted++
		}
	}

	if deleted > 0 {
		js.logger.Info("Cleaned up old jobs", zap.Int("deleted_count", deleted))
	}
}
