package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Gemini    GeminiConfig    `yaml:"gemini"`
	Supabase  SupabaseConfig  `yaml:"supabase"`
	Logger    LoggerConfig    `yaml:"logger"`
	Worker    WorkerConfig    `yaml:"worker"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	OCR       OCRConfig       `yaml:"ocr"`
	Storage   StorageConfig   `yaml:"storage"`
}

type ServerConfig struct {
	Port         int    `yaml:"port"`
	Mode         string `yaml:"mode"`
	APIKey       string `yaml:"api_key"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
	MaxFileSize  int64  `yaml:"max_file_size"`
}

type GeminiConfig struct {
	APIKey     string `yaml:"api_key"`
	Model      string `yaml:"model"`
	Timeout    int    `yaml:"timeout"`
	MaxRetries int    `yaml:"max_retries"`
}

type SupabaseConfig struct {
	URL        string `yaml:"url"`
	Key        string `yaml:"key"`
	Timeout    int    `yaml:"timeout"`
	MaxRetries int    `yaml:"max_retries"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type WorkerConfig struct {
	MaxWorkers int `yaml:"max_workers"`
	QueueSize  int `yaml:"queue_size"`
	JobTimeout int `yaml:"job_timeout"`
}

type RateLimitConfig struct {
	RequestsPerSecond int `yaml:"requests_per_second"`
	Burst             int `yaml:"burst"`
	CleanupInterval   int `yaml:"cleanup_interval"`
}

type OCRConfig struct {
	Enabled       bool   `yaml:"enabled"`
	TesseractPath string `yaml:"tesseract_path"`
	TempDir       string `yaml:"temp_dir"`
}

type StorageConfig struct {
	TempDir         string `yaml:"temp_dir"`
	CleanupInterval int    `yaml:"cleanup_interval"`
	MaxAge          int    `yaml:"max_age"`
}

// Load reads configuration from config.yaml file and environment variables
func Load() (*Config, error) {
	configPath := getConfigPath()
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Replace environment variables in config content
	configContent := expandEnvVars(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(configContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "config/config.yaml"
}

// expandEnvVars replaces ${VAR_NAME} patterns with environment variable values
func expandEnvVars(content string) string {
	return os.Expand(content, func(key string) string {
		return os.Getenv(key)
	})
}

func validate(config *Config) error {
	if config.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}

	if config.Gemini.APIKey == "" {
		return fmt.Errorf("gemini API key is required")
	}

	if config.Supabase.URL == "" {
		return fmt.Errorf("supabase URL is required")
	}

	if config.Supabase.Key == "" {
		return fmt.Errorf("supabase key is required")
	}

	if config.Server.APIKey == "" {
		return fmt.Errorf("server API key is required")
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, config.Logger.Level) {
		return fmt.Errorf("invalid log level: %s", config.Logger.Level)
	}

	// Validate log format
	validFormats := []string{"json", "console"}
	if !contains(validFormats, config.Logger.Format) {
		return fmt.Errorf("invalid log format: %s", config.Logger.Format)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
