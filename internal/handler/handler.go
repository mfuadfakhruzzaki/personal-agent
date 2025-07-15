package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"todo-agent-backend/internal/logger"
	"todo-agent-backend/internal/models"
	"todo-agent-backend/internal/service"
	"todo-agent-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	processingService service.ProcessingServiceInterface
	jobService        service.JobServiceInterface
	logger            *logger.Logger
	apiKey            string
}

func NewHandler(processingService service.ProcessingServiceInterface, jobService service.JobServiceInterface, logger *logger.Logger, apiKey string) *Handler {
	return &Handler{
		processingService: processingService,
		jobService:        jobService,
		logger:            logger,
		apiKey:            apiKey,
	}
}

// HealthCheck handles GET /healthz
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: utils.TimeNow(),
		Version:   "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

// ProcessInput handles POST /process
func (h *Handler) ProcessInput(c *gin.Context) {
	// Authenticate request
	if !h.authenticate(c) {
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		h.logger.Error("Failed to parse multipart form", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse multipart form",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Extract form data
	request := models.ProcessRequest{
		Type:   c.PostForm("type"),
		UserID: c.PostForm("user_id"),
	}

	// Validate required fields
	if request.Type == "" || request.UserID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "type and user_id are required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate type
	if !isValidType(request.Type) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "type must be one of: text, image, document",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Handle different input types
	var content string
	var filePath string

	switch request.Type {
	case "text":
		content = c.PostForm("content")
		if content == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "validation_error",
				Message: "content is required for type 'text'",
				Code:    http.StatusBadRequest,
			})
			return
		}

	case "image", "document":
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "validation_error",
				Message: "file is required for type 'image' or 'document'",
				Code:    http.StatusBadRequest,
			})
			return
		}
		defer file.Close()

		// Validate file
		if err := h.validateFile(header, request.Type); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Save file temporarily
		filePath, err = h.saveUploadedFile(file, header)
		if err != nil {
			h.logger.Error("Failed to save uploaded file", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to save uploaded file",
				Code:    http.StatusInternalServerError,
			})
			return
		}
	}

	// Create job
	job := &models.Job{
		ID:       uuid.New().String(),
		UserID:   request.UserID,
		Type:     request.Type,
		Content:  content,
		FilePath: filePath,
		Status:   models.JobStatusPending,
		CreatedAt: utils.TimeNow(),
		UpdatedAt: utils.TimeNow(),
	}

	// Submit job for processing
	err = h.jobService.SubmitJob(job)
	if err != nil {
		h.logger.Error("Failed to submit job", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to submit job for processing",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Start processing asynchronously
	go h.processingService.ProcessJob(job)

	// Return job ID
	response := models.ProcessResponse{
		JobID:   job.ID,
		Status:  "accepted",
		Message: "Processing request",
	}

	h.logger.Info("Job submitted for processing",
		zap.String("job_id", job.ID),
		zap.String("user_id", job.UserID),
		zap.String("type", job.Type),
	)

	c.JSON(http.StatusAccepted, response)
}

// GetJobStatus handles GET /status/:job_id
func (h *Handler) GetJobStatus(c *gin.Context) {
	// Authenticate request
	if !h.authenticate(c) {
		return
	}

	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "job_id is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get job status
	job, err := h.jobService.GetJob(jobID)
	if err != nil {
		if err == service.ErrJobNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Job not found",
				Code:    http.StatusNotFound,
			})
			return
		}

		h.logger.Error("Failed to get job status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get job status",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Build response
	status := models.JobStatus{
		JobID:     job.ID,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}

	if job.Error != "" {
		status.Message = job.Error
	}

	if job.Result != nil {
		// Convert processing result to todos
		todos := make([]models.Todo, len(job.Result.Todos))
		for i, item := range job.Result.Todos {
			todo := models.Todo{
				ID:          uuid.New(),
				UserID:      job.UserID,
				Title:       item.Title,
				Description: &item.Description,
				SourceType:  job.Type,
				CreatedAt:   job.CreatedAt,
			}

			if item.DueDate != nil && *item.DueDate != "" {
				if dueDate, err := utils.ParseDate(*item.DueDate); err == nil {
					todo.DueDate = &dueDate
				}
			}

			todos[i] = todo
		}
		status.Todos = todos
	}

	c.JSON(http.StatusOK, status)
}

// authenticate validates API key
func (h *Handler) authenticate(c *gin.Context) bool {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		apiKey = c.GetHeader("Authorization")
		if after, ok :=strings.CutPrefix(apiKey, "Bearer "); ok  {
			apiKey = after
		}
	}

	if apiKey != h.apiKey {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing API key",
			Code:    http.StatusUnauthorized,
		})
		return false
	}

	return true
}

// isValidType checks if the input type is valid
func isValidType(inputType string) bool {
	validTypes := []string{"text", "image", "document"}
	for _, t := range validTypes {
		if t == inputType {
			return true
		}
	}
	return false
}

// validateFile validates uploaded file
func (h *Handler) validateFile(header *multipart.FileHeader, inputType string) error {
	// Check file size (5MB max)
	const maxFileSize = 5 * 1024 * 1024
	if header.Size > maxFileSize {
		return fmt.Errorf("file size exceeds 5MB limit")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	
	switch inputType {
	case "image":
		validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
		if !contains(validExts, ext) {
			return fmt.Errorf("invalid image format. Supported: jpg, jpeg, png, gif, bmp, webp")
		}
	case "document":
		validExts := []string{".pdf", ".doc", ".docx", ".txt", ".rtf"}
		if !contains(validExts, ext) {
			return fmt.Errorf("invalid document format. Supported: pdf, doc, docx, txt, rtf")
		}
	}

	return nil
}

// saveUploadedFile saves uploaded file to temporary directory
func (h *Handler) saveUploadedFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Create temp directory if not exists
	tempDir := "/tmp/todo-agent"
	if err := utils.CreateDirIfNotExists(tempDir); err != nil {
		return "", err
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s", uuid.New().String(), header.Filename)
	filePath := filepath.Join(tempDir, filename)

	// Create destination file
	dst, err := utils.CreateFile(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// contains checks if slice contains item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
