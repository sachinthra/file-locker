.PHONY: help dev-setup build run test clean docker-up docker-down

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo '$(BLUE)File Locker - Makefile Commands$(NC)'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

dev-setup: ## Setup development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@chmod +x scripts/setup-dev.sh
	@./scripts/setup-dev.sh

build-backend: ## Build Go backend
	@echo "$(BLUE)Building backend...$(NC)"
	cd backend && go build -o bin/filelocker cmd/server/main.go
	cd backend && go build -o bin/fl cmd/cli/main.go
	@echo "$(GREEN)Backend built successfully!$(NC)"

build-frontend: ## Build Preact frontend
	@echo "$(BLUE)Building frontend...$(NC)"
	cd frontend && npm install && npm run build
	@echo "$(GREEN)Frontend built successfully!$(NC)"

build: build-backend build-frontend ## Build both backend and frontend

install-backend: ## Install Go dependencies
	@echo "$(BLUE)Installing Go dependencies...$(NC)"
	cd backend && go mod download

install-frontend: ## Install Node.js dependencies
	@echo "$(BLUE)Installing Node.js dependencies...$(NC)"
	cd frontend && npm install

install: install-backend install-frontend ## Install all dependencies

test-backend: ## Run Go tests
	@echo "$(BLUE)Running backend tests...$(NC)"
	cd backend && go test ./... -v -race -cover

test-frontend: ## Run frontend tests
	@echo "$(BLUE)Running frontend tests...$(NC)"
	cd frontend && npm test

test: test-backend test-frontend ## Run all tests

lint-backend: ## Lint Go code
	@echo "$(BLUE)Linting backend...$(NC)"
	cd backend && golangci-lint run

lint-frontend: ## Lint frontend code
	@echo "$(BLUE)Linting frontend...$(NC)"
	cd frontend && npm run lint

lint: lint-backend lint-frontend ## Lint all code

run-backend: ## Run backend server
	@echo "$(BLUE)Starting backend server...$(NC)"
	cd backend && go run cmd/server/main.go

run-frontend: ## Run frontend dev server
	@echo "$(BLUE)Starting frontend dev server...$(NC)"
	cd frontend && npm run dev

docker-build: ## Build Docker images
	@echo "$(BLUE)Building Docker images...$(NC)"
	docker-compose build

docker-up: ## Start all services with Docker Compose
	@echo "$(BLUE)Starting services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Services started!$(NC)"
	@echo "Web UI: http://localhost:9010"
	@echo "MinIO Console: http://localhost:9013"

docker-down: ## Stop all services
	@echo "$(YELLOW)Stopping services...$(NC)"
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-restart: docker-down docker-up ## Restart all services

clean-backend: ## Clean backend build artifacts
	@echo "$(YELLOW)Cleaning backend...$(NC)"
	cd backend && rm -rf bin/

clean-frontend: ## Clean frontend build artifacts
	@echo "$(YELLOW)Cleaning frontend...$(NC)"
	cd frontend && rm -rf dist/ node_modules/

clean-docker: ## Remove Docker volumes and images
	@echo "$(RED)Removing Docker volumes and images...$(NC)"
	docker-compose down -v
	docker system prune -f

clean: clean-backend clean-frontend ## Clean all build artifacts

proto-gen: ## Generate gRPC code from proto files
	@echo "$(BLUE)Generating gRPC code...$(NC)"
	cd backend/pkg/proto && protoc --go_out=. --go-grpc_out=. *.proto

format-backend: ## Format Go code
	@echo "$(BLUE)Formatting backend code...$(NC)"
	cd backend && gofmt -w .

format-frontend: ## Format frontend code
	@echo "$(BLUE)Formatting frontend code...$(NC)"
	cd frontend && npm run format

format: format-backend format-frontend ## Format all code

dev: ## Start development environment
	@echo "$(BLUE)Starting development environment...$(NC)"
	@make -j2 run-backend run-frontend

init-minio: ## Initialize MinIO bucket
	@echo "$(BLUE)Initializing MinIO bucket...$(NC)"
	docker-compose exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
	docker-compose exec minio mc mb local/filelocker || true
	@echo "$(GREEN)MinIO bucket initialized!$(NC)"

check-env: ## Check if .env file exists
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Warning: .env file not found. Copying from .env.example...$(NC)"; \
		cp .env.example .env; \
		echo "$(GREEN).env file created. Please update it with your configuration.$(NC)"; \
	fi

status: ## Show service status
	@echo "$(BLUE)Service Status:$(NC)"
	@docker-compose ps

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "$(GREEN)Tools installed!$(NC)"
