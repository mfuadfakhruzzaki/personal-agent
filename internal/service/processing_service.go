package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"todo-agent-backend/internal/logger"
	"todo-agent-backend/internal/models"
	"todo-agent-backend/internal/repository"
	"todo-agent-backend/pkg/gemini"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProcessingService handles the core business logic for processing inputs
type ProcessingService struct {
	geminiClient *gemini.Client
	todoRepo     *repository.TodoRepository
	logger       *logger.Logger
}

// NewProcessingService creates a new processing service
func NewProcessingService(geminiClient *gemini.Client, todoRepo *repository.TodoRepository, logger *logger.Logger) *ProcessingService {
	return &ProcessingService{
		geminiClient: geminiClient,
		todoRepo:     todoRepo,
		logger:       logger,
	}
}

// ProcessJob processes a job asynchronously
func (ps *ProcessingService) ProcessJob(job *models.Job) {
	ps.logger.Info("Starting job processing", zap.String("job_id", job.ID))

	// Update job status to processing
	job.Status = models.JobStatusProcessing
	job.UpdatedAt = time.Now()

	// Extract text content based on job type
	text, err := ps.extractText(job)
	if err != nil {
		ps.logger.Error("Failed to extract text", 
			zap.String("job_id", job.ID), 
			zap.Error(err))
		ps.markJobFailed(job, fmt.Sprintf("Failed to extract text: %v", err))
		return
	}

	// Process with Gemini AI
	todos, err := ps.geminiClient.ExtractTodos(text)
	if err != nil {
		ps.logger.Error("Failed to extract todos with Gemini", 
			zap.String("job_id", job.ID), 
			zap.Error(err))
		ps.markJobFailed(job, fmt.Sprintf("Failed to process with AI: %v", err))
		return
	}

	// Convert to processing result
	result := &models.ProcessingResult{
		Todos:       make([]models.TodoItem, len(todos)),
		ProcessedAt: time.Now(),
	}

	for i, todo := range todos {
		result.Todos[i] = models.TodoItem{
			Title:       todo.Title,
			Description: todo.Description,
			DueDate:     todo.DueDate,
		}
	}

	// Save todos to database
	err = ps.saveTodosToDatabase(job.UserID, result.Todos, job.Type)
	if err != nil {
		ps.logger.Error("Failed to save todos to database", 
			zap.String("job_id", job.ID), 
			zap.Error(err))
		ps.markJobFailed(job, fmt.Sprintf("Failed to save todos: %v", err))
		return
	}

	// Update job with result
	job.Status = models.JobStatusCompleted
	job.Result = result
	job.UpdatedAt = time.Now()

	// Clean up temporary files
	if job.FilePath != "" {
		ps.cleanupFile(job.FilePath)
	}

	ps.logger.Info("Job processing completed", 
		zap.String("job_id", job.ID),
		zap.Int("todos_count", len(result.Todos)))
}

// extractText extracts text content based on job type
func (ps *ProcessingService) extractText(job *models.Job) (string, error) {
	switch job.Type {
	case "text":
		return job.Content, nil
		
	case "image":
		// For images, we'd typically use OCR here
		// For now, we'll pass the file path to Gemini Vision API
		return ps.processImageFile(job.FilePath)
		
	case "document":
		// For documents, we'd parse PDF/DOC files here
		return ps.processDocumentFile(job.FilePath)
		
	default:
		return "", fmt.Errorf("unsupported job type: %s", job.Type)
	}
}

// processImageFile processes image files (placeholder for OCR/Vision API)
func (ps *ProcessingService) processImageFile(filePath string) (string, error) {
	// TODO: Implement image processing with Gemini Vision API
	// For now, return a placeholder
	return fmt.Sprintf("Image content from file: %s", filePath), nil
}

// processDocumentFile processes document files
func (ps *ProcessingService) processDocumentFile(filePath string) (string, error) {
	// Simple text file reading for now
	// TODO: Implement proper PDF/DOC parsing
	if strings.HasSuffix(strings.ToLower(filePath), ".txt") {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read text file: %w", err)
		}
		return string(content), nil
	}
	
	// For other document types, return placeholder
	return fmt.Sprintf("Document content from file: %s", filePath), nil
}

// saveTodosToDatabase saves extracted todos to the database
func (ps *ProcessingService) saveTodosToDatabase(userID string, todoItems []models.TodoItem, sourceType string) error {
	if len(todoItems) == 0 {
		return nil
	}

	todos := make([]models.Todo, len(todoItems))
	now := time.Now()

	for i, item := range todoItems {
		todo := models.Todo{
			ID:          uuid.New(),
			UserID:      userID,
			Title:       item.Title,
			SourceType:  sourceType,
			CreatedAt:   now,
		}

		// Set description if provided
		if item.Description != "" {
			todo.Description = &item.Description
		}

		// Parse due date if provided
		if item.DueDate != nil && *item.DueDate != "" {
			if dueDate, err := time.Parse("2006-01-02", *item.DueDate); err == nil {
				todo.DueDate = &dueDate
			}
		}

		todos[i] = todo
	}

	// Save to database
	return ps.todoRepo.InsertTodos(todos)
}

// markJobFailed marks a job as failed with error message
func (ps *ProcessingService) markJobFailed(job *models.Job, errorMsg string) {
	job.Status = models.JobStatusFailed
	job.Error = errorMsg
	job.UpdatedAt = time.Now()
}

// cleanupFile removes temporary files
func (ps *ProcessingService) cleanupFile(filePath string) {
	if err := os.Remove(filePath); err != nil {
		ps.logger.Warn("Failed to cleanup file", 
			zap.String("file_path", filePath), 
			zap.Error(err))
	}
}
