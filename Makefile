.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down lint

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the service binary
	@echo "Building service..."
	@go build -o ./bin/media_managment_service ./cmd/media_managment_service

build-all: ## Build all binaries
	@echo "Building all binaries..."
	@go build -o ./bin/media_managment_service ./cmd/media_managment_service
	@go build -o ./bin/uploader ./tools/uploader/cmd

# Run targets
run: ## Run the service locally
	@echo "Running service..."
	@go run ./cmd/media_managment_service

# Test targets
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-postgres: ## Run only postgres adapter tests
	@echo "Running postgres tests..."
	@go test -v ./internal/adapters/storage/postgres

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t medias:latest .

docker-up: ## Start all services with docker-compose
	@echo "Starting services..."
	@docker-compose up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down

docker-logs: ## Show docker-compose logs
	@docker-compose logs -f

docker-clean: ## Stop and remove all containers, networks, and volumes
	@echo "Cleaning up Docker resources..."
	@docker-compose down -v

# Database targets
migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@goose -dir ./migrations postgres "$${DB_URL}" up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	@goose -dir ./migrations postgres "$${DB_URL}" down

migrate-status: ## Show migration status
	@goose -dir ./migrations postgres "$${DB_URL}" status

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@goose -dir ./migrations create $(NAME) sql

# Linting and formatting
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

# Clean targets
clean: ## Remove built binaries and temporary files
	@echo "Cleaning..."
	@rm -rf ./bin
	@rm -f coverage.out coverage.html

clean-all: clean docker-clean ## Clean everything including Docker resources

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Local development with DB only
db-up: ## Start only PostgreSQL (for local development)
	@echo "Starting PostgreSQL..."
	@docker-compose up -d postgres

db-down: ## Stop PostgreSQL
	@echo "Stopping PostgreSQL..."
	@docker-compose stop postgres
