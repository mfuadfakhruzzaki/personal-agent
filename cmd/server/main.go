package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-agent-backend/internal/config"
	"todo-agent-backend/internal/handler"
	"todo-agent-backend/internal/logger"
	"todo-agent-backend/internal/middleware"
	"todo-agent-backend/internal/repository"
	"todo-agent-backend/internal/service"
	"todo-agent-backend/pkg/gemini"
	"todo-agent-backend/pkg/supabase"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Format)
	defer logger.Sync()

	logger.Info("Starting Todo Agent Backend Server")

	// Initialize external services
	geminiClient := gemini.NewClient(cfg.Gemini.APIKey, cfg.Gemini.Model)
	supabaseClient := supabase.NewClient(cfg.Supabase.URL, cfg.Supabase.Key)

	// Initialize repository
	todoRepo := repository.NewTodoRepository(supabaseClient)

	// Initialize services
	processingService := service.NewProcessingService(geminiClient, todoRepo, logger)
	jobService := service.NewJobService(logger)

	// Initialize handlers
	handlers := handler.NewHandler(processingService, jobService, logger, cfg.Server.APIKey)

	// Setup Gin router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimit.RequestsPerSecond, 
		cfg.RateLimit.Burst, 
		time.Duration(cfg.RateLimit.CleanupInterval)*time.Second,
	)
	router.Use(rateLimiter.Middleware())

	// Setup routes
	setupRoutes(router, handlers)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server starting on port %d", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(fmt.Sprintf("Failed to start server: %v", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	logger.Info("Server exited")
}

func setupRoutes(router *gin.Engine, h *handler.Handler) {
	// Health check
	router.GET("/healthz", h.HealthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/process", h.ProcessInput)
		api.GET("/status/:job_id", h.GetJobStatus)
	}

	// Backward compatibility - direct routes
	router.POST("/process", h.ProcessInput)
	router.GET("/status/:job_id", h.GetJobStatus)
}
