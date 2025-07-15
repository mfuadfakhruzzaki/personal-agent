package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"todo-agent-backend/internal/logger"
	"todo-agent-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProcessingService for testing
type MockProcessingService struct {
	mock.Mock
}

func (m *MockProcessingService) ProcessJob(job *models.Job) {
	m.Called(job)
}

// MockJobService for testing
type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) SubmitJob(job *models.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobService) GetJob(jobID string) (*models.Job, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobService) UpdateJob(jobID string, status models.JobStatusEnum, result *models.ProcessingResult, errorMsg string) error {
	args := m.Called(jobID, status, result, errorMsg)
	return args.Error(0)
}

func (m *MockJobService) ListJobs(userID string) []*models.Job {
	args := m.Called(userID)
	return args.Get(0).([]*models.Job)
}

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	mockProcessingService := &MockProcessingService{}
	mockJobService := &MockJobService{}
	logger := logger.NewLogger("info", "console")
	
	handler := NewHandler(mockProcessingService, mockJobService, logger, "test-api-key")
	
	router := gin.New()
	router.GET("/healthz", handler.HealthCheck)
	
	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthz", nil)
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
}

func TestProcessInput_Text(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	mockProcessingService := &MockProcessingService{}
	mockJobService := &MockJobService{}
	logger := logger.NewLogger("info", "console")
	
	handler := NewHandler(mockProcessingService, mockJobService, logger, "test-api-key")
	
	// Mock expectations
	mockJobService.On("SubmitJob", mock.AnythingOfType("*models.Job")).Return(nil)
	mockProcessingService.On("ProcessJob", mock.AnythingOfType("*models.Job")).Return()
	
	router := gin.New()
	router.POST("/process", handler.ProcessInput)
	
	// Create form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("type", "text")
	writer.WriteField("content", "Meeting tomorrow at 10am, review code, send report")
	writer.WriteField("user_id", "test-user")
	writer.Close()
	
	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/process", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", "test-api-key")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusAccepted, w.Code)
	
	var response models.ProcessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "accepted", response.Status)
	assert.NotEmpty(t, response.JobID)
	
	// Verify mocks
	mockJobService.AssertExpectations(t)
	mockProcessingService.AssertExpectations(t)
}

func TestProcessInput_InvalidAPIKey(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	mockProcessingService := &MockProcessingService{}
	mockJobService := &MockJobService{}
	logger := logger.NewLogger("info", "console")
	
	handler := NewHandler(mockProcessingService, mockJobService, logger, "test-api-key")
	
	router := gin.New()
	router.POST("/process", handler.ProcessInput)
	
	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/process", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetJobStatus(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	mockProcessingService := &MockProcessingService{}
	mockJobService := &MockJobService{}
	logger := logger.NewLogger("info", "console")
	
	handler := NewHandler(mockProcessingService, mockJobService, logger, "test-api-key")
	
	// Mock job
	job := &models.Job{
		ID:     "test-job-id",
		UserID: "test-user",
		Type:   "text",
		Status: models.JobStatusCompleted,
		Result: &models.ProcessingResult{
			Todos: []models.TodoItem{
				{
					Title:       "Test todo",
					Description: "Test description",
					DueDate:     nil,
				},
			},
		},
	}
	
	mockJobService.On("GetJob", "test-job-id").Return(job, nil)
	
	router := gin.New()
	router.GET("/status/:job_id", handler.GetJobStatus)
	
	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/test-job-id", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.JobStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-job-id", response.JobID)
	assert.Equal(t, "completed", response.Status)
	assert.Len(t, response.Todos, 1)
	
	// Verify mocks
	mockJobService.AssertExpectations(t)
}
