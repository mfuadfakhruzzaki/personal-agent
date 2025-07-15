MAKEFLAGS += --silent

# Variables
APP_NAME = todo-agent
GO_VERSION = 1.21
DOCKER_IMAGE = todo-agent-backend
BUILD_DIR = bin
CONFIG_DIR = config

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Development
.PHONY: dev
dev: ## Run the application in development mode
	@echo "🚀 Starting development server..."
	CONFIG_PATH=config/config.yaml go run cmd/server/main.go

.PHONY: build
build: ## Build the application binary
	@echo "🔨 Building application..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go
	@echo "✅ Build completed: $(BUILD_DIR)/$(APP_NAME)"

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "🔨 Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows go build -a -installsuffix cgo -o $(BUILD_DIR)/$(APP_NAME).exe cmd/server/main.go

.PHONY: build-mac
build-mac: ## Build for macOS
	@echo "🔨 Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo -o $(BUILD_DIR)/$(APP_NAME)-mac cmd/server/main.go

.PHONY: test
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests and show coverage
	@echo "📊 Coverage report:"
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: lint
lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "📝 Formatting code..."
	go fmt ./...
	goimports -w .

.PHONY: deps
deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean

# Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run --rm -p 8080:8080 \
		-e GEMINI_API_KEY="${GEMINI_API_KEY}" \
		-e SUPABASE_URL="${SUPABASE_URL}" \
		-e SUPABASE_KEY="${SUPABASE_KEY}" \
		$(DOCKER_IMAGE):latest

.PHONY: docker-compose-up
docker-compose-up: ## Start services with docker-compose
	@echo "🐳 Starting services with docker-compose..."
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with docker-compose
	@echo "🐳 Stopping services with docker-compose..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## Show docker-compose logs
	docker-compose logs -f

# Configuration
.PHONY: config
config: ## Copy example config
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		echo "📋 Creating config file..."; \
		cp $(CONFIG_DIR)/config.example.yaml $(CONFIG_DIR)/config.yaml; \
		echo "✅ Config file created: $(CONFIG_DIR)/config.yaml"; \
		echo "⚠️  Please edit the config file with your settings"; \
	else \
		echo "📋 Config file already exists"; \
	fi

# Database
.PHONY: db-migrate
db-migrate: ## Run database migrations (placeholder)
	@echo "🗄️ Running database migrations..."
	@echo "ℹ️  Supabase migrations should be run through the Supabase dashboard"

# Deployment
.PHONY: deploy-prepare
deploy-prepare: build ## Prepare deployment files
	@echo "📦 Preparing deployment..."
	@mkdir -p deploy/package
	@cp $(BUILD_DIR)/$(APP_NAME) deploy/package/
	@cp -r $(CONFIG_DIR) deploy/package/
	@cp -r deploy/*.service deploy/package/
	@echo "✅ Deployment package ready in deploy/package/"

.PHONY: deploy-ec2
deploy-ec2: deploy-prepare ## Deploy to EC2 (requires SSH access)
	@echo "🚀 Deploying to EC2..."
	@if [ -z "$(EC2_HOST)" ]; then \
		echo "❌ EC2_HOST environment variable is required"; \
		exit 1; \
	fi
	scp -r deploy/package/* $(EC2_USER)@$(EC2_HOST):/tmp/todo-agent-deploy/
	ssh $(EC2_USER)@$(EC2_HOST) 'sudo /tmp/todo-agent-deploy/deploy.sh'

# Health checks
.PHONY: health
health: ## Check application health
	@echo "🔍 Checking application health..."
	@curl -s http://localhost:8080/healthz || echo "❌ Health check failed"

.PHONY: load-test
load-test: ## Run basic load test
	@echo "🔥 Running load test..."
	@if command -v ab > /dev/null; then \
		ab -n 100 -c 10 http://localhost:8080/healthz; \
	else \
		echo "❌ Apache Bench (ab) not found. Install it to run load tests."; \
	fi

# Monitoring
.PHONY: logs
logs: ## Show application logs (when running with systemd)
	@echo "📜 Showing logs..."
	sudo journalctl -u $(APP_NAME) -f

.PHONY: status
status: ## Show service status
	@echo "📊 Service status:"
	sudo systemctl status $(APP_NAME) --no-pager

# Utilities
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "🛠️ Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: check-env
check-env: ## Check environment variables
	@echo "🔍 Checking environment variables..."
	@echo "GEMINI_API_KEY: $${GEMINI_API_KEY:+set}"
	@echo "SUPABASE_URL: $${SUPABASE_URL:+set}"
	@echo "SUPABASE_KEY: $${SUPABASE_KEY:+set}"
