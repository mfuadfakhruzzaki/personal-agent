package service

import "todo-agent-backend/internal/models"

// ProcessingServiceInterface defines the interface for processing service
type ProcessingServiceInterface interface {
	ProcessJob(job *models.Job)
}

// JobServiceInterface defines the interface for job service
type JobServiceInterface interface {
	SubmitJob(job *models.Job) error
	GetJob(jobID string) (*models.Job, error)
	UpdateJob(jobID string, status models.JobStatusEnum, result *models.ProcessingResult, errorMsg string) error
	ListJobs(userID string) []*models.Job
}
