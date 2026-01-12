.PHONY: help sync install dev dev-local run-backend run-frontend lint format clean build-release

# Colors for pretty printing
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

# Config
DOCKER_USER ?= yourusername
VERSION ?= latest

help: ## Show this help message
	@printf "$(BLUE)File Locker - Project Commander$(NC)\n"
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# -----------------------------------------------------------------
# SETUP & INSTALLATION
# -----------------------------------------------------------------
sync: ## Sync config.yaml to .env (Required for Docker)
	@echo "$(BLUE)Syncing configuration...$(NC)"
	@chmod +x scripts/sync-config.sh
	@./scripts/sync-config.sh

install-backend:
	@echo "$(BLUE)Installing Go dependencies...$(NC)"
	cd backend && go mod download
	@echo "$(GREEN)Go dependencies installed!$(NC)"

install-frontend:
	@echo "$(BLUE)Installing Node.js dependencies...$(NC)"
	cd frontend && npm install
	@echo "$(GREEN)Node.js dependencies installed!$(NC)"

install: install-backend install-frontend ## Install all dependencies

install-tools: ## Install dev tools (golangci-lint, protoc plugins)
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "$(GREEN)Tools installed!$(NC)"

# -----------------------------------------------------------------
# DEVELOPMENT MODES
# -----------------------------------------------------------------
dev: sync ## Full Docker Stack (Everything containerized)
	@echo "$(BLUE)Starting Full Docker Environment...$(NC)"
	DOCKER_DEFAULT_PLATFORM=linux/amd64 docker compose up -d --build
	@echo "$(GREEN)Backend running at http://localhost:9010$(NC)"
	@echo "$(GREEN)Frontend running at http://localhost:80$(NC)"

dev-local: sync ## Hybrid Mode (Docker infra + Native backend/frontend)
	@echo "$(BLUE)Starting Docker Infrastructure...$(NC)"
	docker compose up -d minio redis postgres
	@echo "$(GREEN)Infrastructure started!$(NC)"
	@echo "$(YELLOW)Run these in separate terminals:$(NC)"
	@echo "  make run-backend"
	@echo "  make run-frontend"

run-backend: sync
	@echo "$(BLUE)Starting backend server...$(NC)"
	cd backend && CONFIG_PATH=../configs/config.yaml go run cmd/server/main.go

run-frontend:
	@echo "$(BLUE)Starting frontend dev server...$(NC)"
	cd frontend && npm run dev

# -----------------------------------------------------------------
# DOCKER UTILITIES
# -----------------------------------------------------------------
ps: ## Show Docker service status
	@echo "$(BLUE)Service Status:$(NC)"
	@docker compose ps

logs: ## Follow Docker logs
	docker compose logs -f

down: ## Stop all services
	@echo "$(YELLOW)Stopping services...$(NC)"
	docker compose down

restart: down dev ## Restart all Docker services

# -----------------------------------------------------------------
# QUALITY
# -----------------------------------------------------------------

lint-backend:
	@echo "$(BLUE)Linting backend...$(NC)"
	cd backend && golangci-lint run

lint-frontend:
	@echo "$(BLUE)Linting frontend...$(NC)"
	cd frontend && npm run lint

lint: lint-backend lint-frontend ## Lint all code

format-backend:
	@echo "$(BLUE)Formatting backend...$(NC)"
	cd backend && gofmt -w .

format-frontend:
	@echo "$(BLUE)Formatting frontend...$(NC)"
	cd frontend && npm run format

format: format-backend format-frontend ## Format all code

# -----------------------------------------------------------------
# BUILD & RELEASE
# -----------------------------------------------------------------
build-backend:
	@echo "$(BLUE)Building backend...$(NC)"
	mkdir -p backend/bin
	cd backend && go build -o bin/filelocker cmd/server/main.go
	cd backend && go build -o bin/fl cmd/cli/main.go
	@echo "$(GREEN)Backend built successfully!$(NC)"

build-frontend:
	@echo "$(BLUE)Building frontend...$(NC)"
	cd frontend && npm run build
	@echo "$(GREEN)Frontend built successfully!$(NC)"

build: build-backend build-frontend ## Build both backend and frontend (No Docker)

build-release: ## Build All Artifacts (Docker Images + CLI Binaries + Deb)
	@echo "$(BLUE)Building Release v$(VERSION)...$(NC)"
	@chmod +x scripts/build-release.sh
	@if [ -n "$$DOCKER_USERNAME" ]; then \
		DOCKER_USERNAME=$$DOCKER_USERNAME VERSION=$(VERSION) ./scripts/build-release.sh; \
	else \
		DOCKER_USERNAME=$(DOCKER_USER) VERSION=$(VERSION) ./scripts/build-release.sh; \
	fi

# -----------------------------------------------------------------
# UTILITIES
# -----------------------------------------------------------------
proto-gen: ## Generate gRPC code from proto files
	@echo "$(BLUE)Generating gRPC code...$(NC)"
	cd backend/pkg/proto && protoc --go_out=. --go-grpc_out=. *.proto
	@echo "$(GREEN)Proto files generated!$(NC)"

clean: ## Clean up build artifacts including Docker containers and volumes
	@echo "$(RED)Cleaning artifacts...$(NC)"
	rm -rf bin/ dist/ backend/bin frontend/dist
	docker compose down -v

# -----------------------------------------------------------------
# DISTRIBUTION PACKAGING
# -----------------------------------------------------------------
deb-pi4: ## Build .deb installer for Raspberry Pi 4 (arm64)
	@echo "Packing CLI binary for Raspberry Pi 4..."
	mkdir -p dist/deb/opt/filelocker
	cp bin/fl-linux-arm64 dist/deb/opt/filelocker/fl
	chmod +x dist/deb/opt/filelocker/fl
	mkdir -p dist/deb/DEBIAN
	echo "Package: filelocker\nVersion: 1.0.0\nArchitecture: arm64\nMaintainer: Sachinthra\nDescription: File Locker CLI and deployment scripts for Raspberry Pi 4\n" > dist/deb/DEBIAN/control
	dpkg-deb --build dist/deb dist/filelocker-pi4.deb
	@echo "✅ .deb package built: dist/filelocker-pi4.deb"