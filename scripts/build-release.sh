#!/bin/bash
set -e

# Colors for pretty output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
DOCKER_USERNAME="${DOCKER_USERNAME:-yourusername}"
VERSION="${VERSION:-latest}"
PLATFORMS="linux/amd64,linux/arm64" # For Docker Images

# CLI Cross-Compilation Targets
# Format: "OS/ARCH"
CLI_TARGETS=(
    "darwin/arm64"   # Mac Apple Silicon (Env 3)
    "darwin/amd64"   # Mac Intel (Env 3)
    "windows/amd64"  # Windows (Env 4)
    "linux/amd64"    # Linux Server (Env 2)
    "linux/arm64"    # Raspberry Pi (Env 2)
)

# Validate Docker Hub username
if [ "$DOCKER_USERNAME" == "yourusername" ]; then
    echo -e "${RED}❌ Error: DOCKER_USERNAME is not set!${NC}"
    echo -e "${YELLOW}Please set it before running:${NC}"
    echo -e "  export DOCKER_USERNAME=\"your-dockerhub-username\""
    echo -e "  OR"
    echo -e "  DOCKER_USER=\"your-dockerhub-username\" make build-release"
    exit 1
fi

echo -e "${BLUE}🚀 Starting Release Build [v${VERSION}]${NC}"

# Setup buildx builder (required for multi-platform)
echo -e "${YELLOW}------------------------------------------------${NC}"
echo -e "${BLUE}🔧 Setting up Docker buildx...${NC}"
if ! docker buildx inspect filelocker-builder > /dev/null 2>&1; then
    echo -e "   ${YELLOW}•${NC} Creating buildx builder instance..."
    docker buildx create --name filelocker-builder --use
    echo -e "${GREEN}✅ Builder created${NC}"
else
    echo -e "   ${YELLOW}•${NC} Using existing builder..."
    docker buildx use filelocker-builder
fi

# 1. Build & Push Docker Images (Server - Env 2)
echo -e "${YELLOW}------------------------------------------------${NC}"
echo -e "${BLUE}📦 Building Docker Images (${PLATFORMS})...${NC}"
docker buildx build --platform ${PLATFORMS} \
    -t ${DOCKER_USERNAME}/filelocker-backend:${VERSION} \
    -f backend/Dockerfile backend --push

docker buildx build --platform ${PLATFORMS} \
    -t ${DOCKER_USERNAME}/filelocker-frontend:${VERSION} \
    -f frontend/Dockerfile frontend --push
echo -e "${GREEN}✅ Docker Images Pushed${NC}"

# 2. Build Native CLI Binaries (Client - Env 3 & 4)
echo -e "${YELLOW}------------------------------------------------${NC}"
echo -e "${BLUE}🛠  Compiling CLI Binaries...${NC}"
mkdir -p bin
cd backend
for target in "${CLI_TARGETS[@]}"; do
    # Split "linux/amd64" into GOOS="linux" and GOARCH="amd64"
    IFS='/' read -r os arch <<< "$target"
    
    # Determine output filename (e.g., fl-windows-amd64.exe)
    ext=""
    if [ "$os" == "windows" ]; then ext=".exe"; fi
    output_name="fl-${os}-${arch}${ext}"
    
    echo -e "   ${YELLOW}•${NC} Building ${output_name}..."
    GOOS=$os GOARCH=$arch go build \
        -ldflags="-s -w -X main.Version=${VERSION}" \
        -o "../bin/${output_name}" \
        ./cmd/cli/main.go
done
cd ..
echo -e "${GREEN}✅ CLI Binaries Built in ./bin/${NC}"

# 3. Package Debian Installer (Ubuntu/Pi - Env 2)
echo -e "${YELLOW}------------------------------------------------${NC}"
echo -e "${BLUE}📦 Packaging .deb Installer...${NC}"
mkdir -p dist/deb/opt/filelocker

# Copy files and replace DOCKER_USERNAME and VERSION placeholders
echo -e "   ${YELLOW}•${NC} Injecting Docker username: ${DOCKER_USERNAME}"
echo -e "   ${YELLOW}•${NC} Injecting version: ${VERSION}"
sed -e "s/\${DOCKER_USERNAME}/${DOCKER_USERNAME}/g" \
    -e "s/:latest/:${VERSION}/g" \
    install/docker-compose.yml > dist/deb/opt/filelocker/docker-compose.yml
sed "s/DOCKER_USERNAME=_YOUR_DOCKER_USER_NAME/DOCKER_USERNAME=${DOCKER_USERNAME}/g" install/setup.sh > dist/deb/opt/filelocker/setup.sh
chmod +x dist/deb/opt/filelocker/setup.sh
cp configs/config.yaml dist/deb/opt/filelocker/

echo -e "${GREEN}✅ Debian Structure created in ./dist/deb/${NC}"
echo -e "   ${YELLOW}(Ready for dpkg-deb build)${NC}"
echo ""
echo -e "${GREEN}🎉 Release Complete!${NC}"
