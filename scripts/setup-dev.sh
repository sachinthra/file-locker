#!/bin/bash

# File Locker Development Setup Script

set -e

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}==================================${NC}"
echo -e "${BLUE}File Locker - Development Setup${NC}"
echo -e "${BLUE}==================================${NC}"
echo ""

# Check if required commands exist
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check Go
echo -e "${BLUE}Checking Go installation...${NC}"
if ! command_exists go; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.21 or higher.${NC}"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓ Go found: ${GO_VERSION}${NC}"

# Check Node.js
echo -e "${BLUE}Checking Node.js installation...${NC}"
if ! command_exists node; then
    echo -e "${RED}Error: Node.js is not installed. Please install Node.js 18 or higher.${NC}"
    exit 1
fi
NODE_VERSION=$(node --version)
echo -e "${GREEN}✓ Node.js found: ${NODE_VERSION}${NC}"

# Check Docker
echo -e "${BLUE}Checking Docker installation...${NC}"
if ! command_exists docker; then
    echo -e "${YELLOW}Warning: Docker is not installed. Docker is required for deployment.${NC}"
else
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | tr -d ',')
    echo -e "${GREEN}✓ Docker found: ${DOCKER_VERSION}${NC}"
fi

# Check Docker Compose
echo -e "${BLUE}Checking Docker Compose installation...${NC}"
if ! command_exists docker-compose; then
    echo -e "${YELLOW}Warning: Docker Compose is not installed. Docker Compose is required for deployment.${NC}"
else
    COMPOSE_VERSION=$(docker-compose --version | awk '{print $4}' | tr -d ',')
    echo -e "${GREEN}✓ Docker Compose found: ${COMPOSE_VERSION}${NC}"
fi

echo ""
echo -e "${BLUE}Setting up environment...${NC}"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from .env.example...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✓ .env file created${NC}"
    echo -e "${YELLOW}⚠ Please update .env with your configuration before running!${NC}"
else
    echo -e "${GREEN}✓ .env file already exists${NC}"
fi

# Create necessary directories
echo -e "${BLUE}Creating project directories...${NC}"
mkdir -p backend/{cmd/{server,cli},internal/{crypto,api,grpc,storage,auth,worker,models},pkg/proto,configs,scripts}
mkdir -p frontend/{src/{components,hooks,services,pages,utils},public}
mkdir -p docs
echo -e "${GREEN}✓ Directories created${NC}"

# Install backend dependencies
echo ""
echo -e "${BLUE}Installing backend dependencies...${NC}"
cd backend
if [ ! -f go.mod ]; then
    echo -e "${YELLOW}Initializing Go module...${NC}"
    go mod init github.com/[username]/file-locker
fi
go mod download
cd ..
echo -e "${GREEN}✓ Backend dependencies installed${NC}"

# Install frontend dependencies
echo ""
echo -e "${BLUE}Installing frontend dependencies...${NC}"
cd frontend
if [ ! -f package.json ]; then
    echo -e "${YELLOW}package.json not found. You'll need to create it manually.${NC}"
else
    npm install
    echo -e "${GREEN}✓ Frontend dependencies installed${NC}"
fi
cd ..

# Install development tools (optional)
echo ""
echo -e "${BLUE}Installing optional development tools...${NC}"
if command_exists go; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || echo -e "${YELLOW}Warning: Could not install golangci-lint${NC}"
    
    echo "Installing protoc generators..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest || echo -e "${YELLOW}Warning: Could not install protoc-gen-go${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest || echo -e "${YELLOW}Warning: Could not install protoc-gen-go-grpc${NC}"
fi

echo ""
echo -e "${GREEN}==================================${NC}"
echo -e "${GREEN}Setup Complete!${NC}"
echo -e "${GREEN}==================================${NC}"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo -e "1. Update .env file with your configuration"
echo -e "2. Start services: ${GREEN}make docker-up${NC} or ${GREEN}docker-compose up -d${NC}"
echo -e "3. Initialize MinIO bucket: ${GREEN}make init-minio${NC}"
echo -e "4. Start development:"
echo -e "   - Backend: ${GREEN}make run-backend${NC}"
echo -e "   - Frontend: ${GREEN}make run-frontend${NC}"
echo -e "   - Both: ${GREEN}make dev${NC}"
echo ""
echo -e "For more commands, run: ${GREEN}make help${NC}"
echo ""
