#!/bin/bash
# ================================================================
# File Locker - Release Build Script
# Builds multi-platform Docker images and native CLI binaries
# ================================================================

set -e  # Exit on error

# ----------------------------------------------------------------
# Configuration
# ----------------------------------------------------------------
DOCKER_USERNAME="${DOCKER_USERNAME:-youruser}"
VERSION="${VERSION:-latest}"
BUILD_DIR="$(pwd)"
BIN_DIR="${BUILD_DIR}/bin"
DIST_DIR="${BUILD_DIR}/dist"
DEB_DIR="${DIST_DIR}/deb"

# Platforms for Docker images
DOCKER_PLATFORMS="linux/amd64,linux/arm64"

# CLI build targets
CLI_TARGETS=(
    "darwin/arm64"   # Mac M1/M2/M3
    "darwin/amd64"   # Mac Intel
    "linux/amd64"    # Linux x86_64
    "linux/arm64"    # Linux ARM64 (Raspberry Pi)
    "windows/amd64"  # Windows
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ----------------------------------------------------------------
# Helper Functions
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

log_step() {
    echo -e "${BLUE}==>${NC} $1"
}

check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 is not installed or not in PATH"
        return 1
    fi
    return 0
}

# ----------------------------------------------------------------
# Pre-flight Checks
# ----------------------------------------------------------------

log_step "Running pre-flight checks..."

# Check Docker
if ! check_command docker; then
    log_error "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check Docker Buildx
if ! docker buildx version &> /dev/null; then
    log_error "Docker Buildx is required but not available"
    log_error "Please update Docker Desktop to the latest version"
    exit 1
fi

# Check Go
if ! check_command go; then
    log_error "Go is not installed"
    log_error "Please install Go: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
log_info "Go version: $GO_VERSION"

# Check Docker login
if ! docker info | grep -q "Username:"; then
    log_warn "Not logged in to Docker Hub"
    read -p "Do you want to login now? (Y/n): " DO_LOGIN
    if [[ -z "$DO_LOGIN" ]] || [[ "$DO_LOGIN" =~ ^[Yy]$ ]]; then
        docker login
    else
        log_error "Docker login required to push images"
        exit 1
    fi
fi

log_info "Pre-flight checks passed ✅"

# ----------------------------------------------------------------
# Create Build Directories
# ----------------------------------------------------------------

log_step "Creating build directories..."

mkdir -p "$BIN_DIR"
mkdir -p "$DEB_DIR/opt/filelocker"
mkdir -p "$DEB_DIR/DEBIAN"

log_info "Build directories created"

# ----------------------------------------------------------------
# Build Docker Images
# ----------------------------------------------------------------

log_step "Building Docker images for: $DOCKER_PLATFORMS"

# Create/use buildx builder
if ! docker buildx ls | grep -q "multiplatform"; then
    log_info "Creating multi-platform builder..."
    docker buildx create --name multiplatform --use
    docker buildx inspect --bootstrap
else
    log_info "Using existing multi-platform builder..."
    docker buildx use multiplatform
fi

# Build Backend Image
log_info "Building backend image..."
docker buildx build \
    --platform ${DOCKER_PLATFORMS} \
    --tag ${DOCKER_USERNAME}/filelocker-backend:${VERSION} \
    --tag ${DOCKER_USERNAME}/filelocker-backend:$(date +%Y%m%d) \
    --file ./backend/Dockerfile \
    --push \
    ./backend

if [ $? -eq 0 ]; then
    log_info "✅ Backend image built successfully!"
else
    log_error "❌ Backend image build failed!"
    exit 1
fi

# Build Frontend Image
log_info "Building frontend image..."
docker buildx build \
    --platform ${DOCKER_PLATFORMS} \
    --tag ${DOCKER_USERNAME}/filelocker-frontend:${VERSION} \
    --tag ${DOCKER_USERNAME}/filelocker-frontend:$(date +%Y%m%d) \
    --file ./frontend/Dockerfile \
    --push \
    ./frontend

if [ $? -eq 0 ]; then
    log_info "✅ Frontend image built successfully!"
else
    log_error "❌ Frontend image build failed!"
    exit 1
fi

# ----------------------------------------------------------------
# Build CLI Binaries
# ----------------------------------------------------------------

log_step "Building CLI binaries..."

cd backend

for target in "${CLI_TARGETS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$target"
    
    OUTPUT_NAME="fl-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    log_info "Building $OUTPUT_NAME..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w -X main.Version=${VERSION}" \
        -o "../${BIN_DIR}/${OUTPUT_NAME}" \
        ./cmd/cli/main.go
    
    if [ $? -eq 0 ]; then
        FILE_SIZE=$(du -h "../${BIN_DIR}/${OUTPUT_NAME}" | cut -f1)
        log_info "✅ Built $OUTPUT_NAME (${FILE_SIZE})"
    else
        log_error "❌ Failed to build $OUTPUT_NAME"
        exit 1
    fi
done

cd ..

# ----------------------------------------------------------------
# Create Debian Package Structure
# ----------------------------------------------------------------

log_step "Creating Debian package structure..."

# Copy installation files
cp install/docker-compose.yml "$DEB_DIR/opt/filelocker/"
cp install/setup.sh "$DEB_DIR/opt/filelocker/"
cp configs/config.yaml "$DEB_DIR/opt/filelocker/"

# Create postinstall script
cat > "$DEB_DIR/opt/filelocker/README.txt" << 'EOF'
File Locker - Installation Instructions
========================================

This package contains the File Locker server installation files.

Installation Steps:
1. cd /opt/filelocker
2. ./setup.sh
3. Follow the interactive prompts

The setup script will:
- Detect Docker installation
- Generate secure passwords
- Configure the web port
- Create a .env file
- Optionally start the services

For more information, visit: https://github.com/youruser/filelocker
EOF

# Create control file for .deb package
cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: filelocker
Version: 1.0.0
Section: web
Priority: optional
Architecture: all
Depends: docker.io | docker-ce
Maintainer: Your Name <your.email@example.com>
Description: File Locker - Secure file sharing and storage
 File Locker is a self-hosted secure file sharing platform
 with end-to-end encryption and expiring links.
 .
 This package installs the Docker Compose configuration
 and setup scripts for easy deployment.
EOF

chmod 755 "$DEB_DIR/opt/filelocker/setup.sh"
chmod 644 "$DEB_DIR/opt/filelocker/docker-compose.yml"
chmod 644 "$DEB_DIR/opt/filelocker/config.yaml"
chmod 644 "$DEB_DIR/opt/filelocker/README.txt"

log_info "Debian package structure created at: $DEB_DIR"

# ----------------------------------------------------------------
# Create Release Archive
# ----------------------------------------------------------------

log_step "Creating release archives..."

# Archive for server deployment
cd "$DIST_DIR"
tar -czf "filelocker-server-${VERSION}.tar.gz" -C deb opt DEBIAN
log_info "Created: filelocker-server-${VERSION}.tar.gz"

# Archive for CLI binaries
cd "$BUILD_DIR"
tar -czf "${DIST_DIR}/filelocker-cli-${VERSION}.tar.gz" -C bin .
log_info "Created: filelocker-cli-${VERSION}.tar.gz"

cd "$BUILD_DIR"

# ----------------------------------------------------------------
# Build Summary
# ----------------------------------------------------------------

echo ""
echo "================================================================"
log_info "Build Complete! 🎉"
echo "================================================================"
echo ""
echo "📦 Docker Images (pushed to Docker Hub):"
echo "   • ${DOCKER_USERNAME}/filelocker-backend:${VERSION}"
echo "   • ${DOCKER_USERNAME}/filelocker-frontend:${VERSION}"
echo ""
echo "💻 CLI Binaries (in bin/):"
for target in "${CLI_TARGETS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$target"
    OUTPUT_NAME="fl-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    if [ -f "${BIN_DIR}/${OUTPUT_NAME}" ]; then
        FILE_SIZE=$(du -h "${BIN_DIR}/${OUTPUT_NAME}" | cut -f1)
        echo "   • ${OUTPUT_NAME} (${FILE_SIZE})"
    fi
done
echo ""
echo "📁 Debian Package Structure:"
echo "   • ${DEB_DIR}/"
echo ""
echo "📦 Release Archives:"
echo "   • ${DIST_DIR}/filelocker-server-${VERSION}.tar.gz"
echo "   • ${DIST_DIR}/filelocker-cli-${VERSION}.tar.gz"
echo ""
echo "================================================================"
echo ""
echo "Next Steps:"
echo ""
echo "1. 🏗️  Build .deb package (requires dpkg-deb):"
echo "   dpkg-deb --build ${DEB_DIR} ${DIST_DIR}/filelocker-${VERSION}.deb"
echo ""
echo "2. 🚀 Server Deployment:"
echo "   • Copy dist/deb/opt/filelocker/ to your server"
echo "   • Run ./setup.sh"
echo "   • Start with docker compose up -d"
echo ""
echo "3. 💻 CLI Distribution:"
echo "   • Mac (ARM): bin/fl-darwin-arm64"
echo "   • Mac (Intel): bin/fl-darwin-amd64"
echo "   • Windows: bin/fl-windows-amd64.exe"
echo "   • Linux: bin/fl-linux-amd64"
echo ""
echo "4. 🔐 CLI Usage:"
echo "   fl login --host http://your-server:8080 -u admin -p password123"
echo "   fl upload myfile.txt"
echo "   fl ls"
echo ""
echo "================================================================"
