#!/bin/bash
# ================================================================
# Build and Push Docker Images for Raspberry Pi (ARM64)
# Run this on your development machine (x86/amd64)
# ================================================================

set -e  # Exit on error

# ----------------------------------------------------------------
# Configuration
# ----------------------------------------------------------------
DOCKER_USERNAME="${DOCKER_USERNAME:-youruser}"
BACKEND_IMAGE="${DOCKER_USERNAME}/filelocker-backend"
FRONTEND_IMAGE="${DOCKER_USERNAME}/filelocker-frontend"
VERSION="${VERSION:-latest}"
PLATFORM="linux/arm64"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ----------------------------------------------------------------
# Functions
# ----------------------------------------------------------------
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# ----------------------------------------------------------------
# Pre-flight Checks
# ----------------------------------------------------------------
log_info "Starting Raspberry Pi ARM64 image build process..."

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker Desktop."
    exit 1
fi

# Check if buildx is available
if ! docker buildx version >/dev/null 2>&1; then
    log_error "Docker buildx is not available. Please update Docker Desktop."
    exit 1
fi

# Check if user is logged in to Docker Hub
if ! docker info | grep -q "Username:"; then
    log_warn "Not logged in to Docker Hub. Running 'docker login'..."
    docker login
fi

# Create/use buildx builder for multi-platform builds
if ! docker buildx ls | grep -q "multiplatform"; then
    log_info "Creating multi-platform builder..."
    docker buildx create --name multiplatform --use
    docker buildx inspect --bootstrap
else
    log_info "Using existing multi-platform builder..."
    docker buildx use multiplatform
fi

# ----------------------------------------------------------------
# Build Backend Image
# ----------------------------------------------------------------
log_info "Building backend image for ARM64..."
docker buildx build \
    --platform ${PLATFORM} \
    --tag ${BACKEND_IMAGE}:${VERSION} \
    --tag ${BACKEND_IMAGE}:$(date +%Y%m%d) \
    --file ./backend/Dockerfile \
    --push \
    ./backend

if [ $? -eq 0 ]; then
    log_info "✅ Backend image built and pushed successfully!"
    log_info "   Image: ${BACKEND_IMAGE}:${VERSION}"
else
    log_error "❌ Backend build failed!"
    exit 1
fi

# ----------------------------------------------------------------
# Build Frontend Image
# ----------------------------------------------------------------
log_info "Building frontend image for ARM64..."
docker buildx build \
    --platform ${PLATFORM} \
    --tag ${FRONTEND_IMAGE}:${VERSION} \
    --tag ${FRONTEND_IMAGE}:$(date +%Y%m%d) \
    --file ./frontend/Dockerfile \
    --push \
    ./frontend

if [ $? -eq 0 ]; then
    log_info "✅ Frontend image built and pushed successfully!"
    log_info "   Image: ${FRONTEND_IMAGE}:${VERSION}"
else
    log_error "❌ Frontend build failed!"
    exit 1
fi

# ----------------------------------------------------------------
# Summary
# ----------------------------------------------------------------
echo ""
log_info "================================================"
log_info "Build Complete! 🎉"
log_info "================================================"
echo ""
log_info "Images pushed to Docker Hub:"
echo "  📦 ${BACKEND_IMAGE}:${VERSION}"
echo "  📦 ${FRONTEND_IMAGE}:${VERSION}"
echo ""
log_info "Next steps:"
echo "  1. Update docker-compose.prod.yml with your Docker Hub username"
echo "  2. Copy to Raspberry Pi:"
echo "     - docker-compose.prod.yml"
echo "     - .env"
echo "     - configs/config.yaml"
echo "     - backend/init-db/"
echo ""
echo "  3. On Raspberry Pi, run:"
echo "     docker compose -f docker-compose.prod.yml pull"
echo "     docker compose -f docker-compose.prod.yml up -d"
echo ""
log_info "================================================"
