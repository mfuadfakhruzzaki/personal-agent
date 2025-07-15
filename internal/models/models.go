package models

import (
	"time"

	"github.com/google/uuid"
)

// Todo represents a todo item in the database
type Todo struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description" db:"description"`
	DueDate     *time.Time `json:"due_date" db:"due_date"`
	SourceType  string     `json:"source_type" db:"source_type"`
	SourceURL   *string    `json:"source_url" db:"source_url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// ProcessRequest represents the input request for processing
type ProcessRequest struct {
	Type    string `form:"type" binding:"required,oneof=text image document"`
	Content string `form:"content"`
	UserID  string `form:"user_id" binding:"required"`
}

// ProcessResponse represents the response from processing endpoint
type ProcessResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// JobStatus represents the status of a processing job
type JobStatus struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Todos     []Todo    `json:"todos,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Job represents a processing job
type Job struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	Type      string             `json:"type"`
	Content   string             `json:"content"`
	FilePath  string             `json:"file_path,omitempty"`
	Status    JobStatusEnum      `json:"status"`
	Result    *ProcessingResult  `json:"result,omitempty"`
	Error     string             `json:"error,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// JobStatusEnum represents possible job statuses
type JobStatusEnum string

const (
	JobStatusPending    JobStatusEnum = "pending"
	JobStatusProcessing JobStatusEnum = "processing"
	JobStatusCompleted  JobStatusEnum = "completed"
	JobStatusFailed     JobStatusEnum = "failed"
)

// ProcessingResult represents the result of AI processing
type ProcessingResult struct {
	Todos       []TodoItem `json:"todos"`
	ProcessedAt time.Time  `json:"processed_at"`
}

// TodoItem represents a todo item from AI processing
type TodoItem struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"` // YYYY-MM-DD format or null
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// ErrorResponse represents error response format
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
