.PHONY: help build run test test-coverage clean docker-build docker-run docker-stop proto vendor lint format

# Variables
APP_NAME=filelocker
VERSION?=1.0.0
BUILD_DIR=bin
DOCKER_IMAGE=$(APP_NAME):$(VERSION)
DOCKER_IMAGE_LATEST=$(APP_NAME):latest
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
PROTO_FILES=$(shell find pkg/proto -name '*.proto')

# Build info
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Colors for output
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

## help: Show this help message
help:
	@echo '${GREEN}File Locker Backend - Makefile Commands${NC}'
	@echo ''
	@echo '${BLUE}Development:${NC}'
	@echo '  make run              - Run the application'
	@echo '  make build            - Build the application binary'
	@echo '  make test             - Run all tests'
	@echo '  make test-coverage    - Run tests with coverage report'
	@echo '  make lint             - Run linters'
	@echo '  make format           - Format Go code'
	@echo ''
	@echo '${BLUE}Docker:${NC}'
	@echo '  make docker-build     - Build Docker image'
	@echo '  make docker-run       - Run Docker container'
	@echo '  make docker-stop      - Stop Docker container'
	@echo '  make docker-compose-up   - Start all services with docker-compose'
	@echo '  make docker-compose-down - Stop all services'
	@echo ''
	@echo '${BLUE}Protobuf:${NC}'
	@echo '  make proto            - Generate protobuf code'
	@echo ''
	@echo '${BLUE}Utilities:${NC}'
	@echo '  make clean            - Clean build artifacts'
	@echo '  make vendor           - Vendor dependencies'
	@echo '  make deps             - Download dependencies'
	@echo ''

## build: Build the application binary
build:
	@echo "${BLUE}Building $(APP_NAME)...${NC}"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go
	@echo "${GREEN}✓ Build complete: $(BUILD_DIR)/$(APP_NAME)${NC}"

## build-linux: Build for Linux (useful for Docker)
build-linux:
	@echo "${BLUE}Building $(APP_NAME) for Linux...${NC}"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux cmd/server/main.go
	@echo "${GREEN}✓ Linux build complete${NC}"

## run: Run the application
run:
	@echo "${BLUE}Starting $(APP_NAME)...${NC}"
	@go run cmd/server/main.go

## deps: Download and install dependencies
deps:
	@echo "${BLUE}Downloading dependencies...${NC}"
	@go mod download
	@echo "${GREEN}✓ Dependencies downloaded${NC}"

## vendor: Vendor dependencies
vendor:
	@echo "${BLUE}Vendoring dependencies...${NC}"
	@go mod vendor
	@echo "${GREEN}✓ Dependencies vendored${NC}"

## test: Run all tests
test:
	@echo "${BLUE}Running tests...${NC}"
	@go test ./... -v -race
	@echo "${GREEN}✓ Tests complete${NC}"

## test-short: Run short tests only
test-short:
	@echo "${BLUE}Running short tests...${NC}"
	@go test ./... -short -v
	@echo "${GREEN}✓ Short tests complete${NC}"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "${BLUE}Running tests with coverage...${NC}"
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}✓ Coverage report generated: coverage.html${NC}"
	@go tool cover -func=coverage.out | grep total:

## bench: Run benchmarks
bench:
	@echo "${BLUE}Running benchmarks...${NC}"
	@go test ./... -bench=. -benchmem

## lint: Run linters
lint:
	@echo "${BLUE}Running linters...${NC}"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "${RED}golangci-lint not installed. Install: https://golangci-lint.run/usage/install/${NC}"; \
		go vet ./...; \
	fi
	@echo "${GREEN}✓ Linting complete${NC}"

## format: Format Go code
format:
	@echo "${BLUE}Formatting code...${NC}"
	@gofmt -s -w $(GO_FILES)
	@echo "${GREEN}✓ Code formatted${NC}"

## proto: Generate protobuf code
proto:
	@echo "${BLUE}Generating protobuf code...${NC}"
	@if command -v protoc >/dev/null 2>&1; then \
		cd pkg/proto && protoc --go_out=. --go-grpc_out=. $(notdir $(PROTO_FILES)); \
		echo "${GREEN}✓ Protobuf code generated${NC}"; \
	else \
		echo "${RED}protoc not installed. Install: https://grpc.io/docs/protoc-installation/${NC}"; \
		exit 1; \
	fi

## docker-build: Build Docker image
docker-build:
	@echo "${BLUE}Building Docker image: $(DOCKER_IMAGE)${NC}"
	@docker build -t $(DOCKER_IMAGE) -t $(DOCKER_IMAGE_LATEST) .
	@echo "${GREEN}✓ Docker image built: $(DOCKER_IMAGE)${NC}"

## docker-run: Run Docker container
docker-run:
	@echo "${BLUE}Running Docker container...${NC}"
	@docker run -d \
		--name $(APP_NAME) \
		-p 9010:9010 \
		-p 9011:9011 \
		-v $(PWD)/configs:/root/configs:ro \
		--link minio:minio \
		--link redis:redis \
		$(DOCKER_IMAGE_LATEST)
	@echo "${GREEN}✓ Container started: $(APP_NAME)${NC}"
	@docker ps | grep $(APP_NAME)

## docker-stop: Stop and remove Docker container
docker-stop:
	@echo "${BLUE}Stopping Docker container...${NC}"
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true
	@echo "${GREEN}✓ Container stopped${NC}"

## docker-logs: Show Docker container logs
docker-logs:
	@docker logs -f $(APP_NAME)

## docker-compose-up: Start all services with docker-compose
docker-compose-up:
	@echo "${BLUE}Starting services with docker-compose...${NC}"
	@cd .. && docker-compose up -d
	@echo "${GREEN}✓ Services started${NC}"
	@cd .. && docker-compose ps

## docker-compose-down: Stop all services
docker-compose-down:
	@echo "${BLUE}Stopping services...${NC}"
	@cd .. && docker-compose down
	@echo "${GREEN}✓ Services stopped${NC}"

## docker-compose-logs: Show docker-compose logs
docker-compose-logs:
	@cd .. && docker-compose logs -f backend

## docker-compose-restart: Restart backend service
docker-compose-restart:
	@echo "${BLUE}Restarting backend service...${NC}"
	@cd .. && docker-compose restart backend
	@echo "${GREEN}✓ Backend restarted${NC}"

## docker-compose-rebuild: Rebuild and restart backend
docker-compose-rebuild:
	@echo "${BLUE}Rebuilding backend service...${NC}"
	@cd .. && docker-compose up -d --build backend
	@echo "${GREEN}✓ Backend rebuilt and restarted${NC}"

## clean: Clean build artifacts and test caches
clean:
	@echo "${BLUE}Cleaning...${NC}"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean -testcache
	@echo "${GREEN}✓ Cleaned${NC}"

## install: Install the binary to GOPATH/bin
install: build
	@echo "${BLUE}Installing $(APP_NAME)...${NC}"
	@cp $(BUILD_DIR)/$(APP_NAME) $(GOPATH)/bin/
	@echo "${GREEN}✓ Installed to $(GOPATH)/bin/$(APP_NAME)${NC}"

## dev-setup: Setup development environment
dev-setup:
	@echo "${BLUE}Setting up development environment...${NC}"
	@echo "Installing dependencies..."
	@go mod download
	@echo "Starting Docker services..."
	@cd .. && docker-compose up -d minio redis
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "${GREEN}✓ Development environment ready${NC}"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Copy configs/config.yaml.example to configs/config.yaml (if needed)"
	@echo "  2. Run: make run"

## dev-stop: Stop development services
dev-stop:
	@echo "${BLUE}Stopping development services...${NC}"
	@cd .. && docker-compose stop minio redis
	@echo "${GREEN}✓ Development services stopped${NC}"

## check: Run all checks (format, lint, test)
check: format lint test
	@echo "${GREEN}✓ All checks passed${NC}"

## ci: Run CI pipeline (no formatting)
ci: lint test
	@echo "${GREEN}✓ CI pipeline complete${NC}"

## security: Run security checks
security:
	@echo "${BLUE}Running security checks...${NC}"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "${RED}gosec not installed. Install: go install github.com/securego/gosec/v2/cmd/gosec@latest${NC}"; \
	fi

## mod-tidy: Tidy go.mod
mod-tidy:
	@echo "${BLUE}Tidying go.mod...${NC}"
	@go mod tidy
	@echo "${GREEN}✓ go.mod tidied${NC}"

## version: Show version information
version:
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

## all: Build everything
all: clean proto build test
	@echo "${GREEN}✓ All tasks complete${NC}"

.DEFAULT_GOAL := help
