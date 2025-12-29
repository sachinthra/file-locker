#!/bin/bash
# ================================================================
# File Locker - Interactive Setup Script
# Generates .env file with secure defaults and user preferences
# ================================================================

set -e  # Exit on error

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

generate_password() {
    # Generate 32 character random password
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

check_port() {
    local port=$1
    # Check if port is in use (works on both Linux and macOS)
    if command -v lsof &> /dev/null; then
        lsof -i:$port &> /dev/null && return 1 || return 0
    elif command -v netstat &> /dev/null; then
        netstat -an | grep ":$port " | grep LISTEN &> /dev/null && return 1 || return 0
    else
        # If neither command available, assume port is free
        return 0
    fi
}

# ----------------------------------------------------------------
# Detect Docker Command
# ----------------------------------------------------------------

log_step "Detecting Docker installation..."

DOCKER_CMD=""
if command -v docker &> /dev/null; then
    # Check if "docker compose" (V2) works
    if docker compose version &> /dev/null; then
        DOCKER_CMD="docker compose"
        log_info "Found Docker Compose V2"
    # Fall back to "docker-compose" (V1)
    elif command -v docker-compose &> /dev/null; then
        DOCKER_CMD="docker-compose"
        log_info "Found Docker Compose V1"
    else
        log_error "Docker is installed but Docker Compose is not available!"
        log_error "Please install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
else
    log_error "Docker is not installed!"
    log_error "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

log_info "Using command: ${DOCKER_CMD}"

# ----------------------------------------------------------------
# Docker Hub Username (Pre-configured during build)
# ----------------------------------------------------------------

DOCKER_USERNAME=_YOUR_DOCKER_USER_NAME  # This gets replaced during build-release

log_info "Using Docker images: ${DOCKER_USERNAME}/filelocker-*:latest"

# ----------------------------------------------------------------
# Port Selection
# ----------------------------------------------------------------

log_step "Port Configuration"
echo ""

DEFAULT_PORT=8080
WEB_PORT=$DEFAULT_PORT

log_info "The web interface needs a port to listen on (default: $DEFAULT_PORT)"

while true; do
    if check_port $WEB_PORT; then
        log_info "Port $WEB_PORT is available!"
        read -p "Use port $WEB_PORT? (Y/n): " USE_PORT
        if [[ -z "$USE_PORT" ]] || [[ "$USE_PORT" =~ ^[Yy]$ ]]; then
            break
        else
            read -p "Enter a different port: " WEB_PORT
        fi
    else
        log_warn "Port $WEB_PORT is already in use!"
        read -p "Enter a different port: " WEB_PORT
    fi
done

log_info "Web interface will be available at: http://localhost:$WEB_PORT"

# ----------------------------------------------------------------
# Generate Secrets
# ----------------------------------------------------------------

log_step "Generating secure passwords..."

DB_PASSWORD=$(generate_password)
MINIO_SECRET=$(generate_password)
JWT_SECRET=$(generate_password)

log_info "Generated secure random passwords"

# ----------------------------------------------------------------
# Write .env File
# ----------------------------------------------------------------

ENV_FILE=".env"

log_step "Writing configuration to $ENV_FILE..."

cat > "$ENV_FILE" << EOF
# ================================================================
# File Locker - Environment Configuration
# Generated on: $(date)
# ================================================================

# ----------------------------------------------------------------
# Docker Hub Configuration
# ----------------------------------------------------------------
DOCKER_USERNAME=$DOCKER_USERNAME

# ----------------------------------------------------------------
# Web Interface Port
# ----------------------------------------------------------------
WEB_PORT=$WEB_PORT

# ----------------------------------------------------------------
# Database Configuration
# ----------------------------------------------------------------
DB_USER=filelocker_user
DB_PASSWORD=$DB_PASSWORD
DB_NAME=filelocker_db

# ----------------------------------------------------------------
# MinIO Object Storage
# ----------------------------------------------------------------
MINIO_ACCESS_KEY=filelocker_minio
MINIO_SECRET_KEY=$MINIO_SECRET
MINIO_BUCKET=filelocker-files

# ----------------------------------------------------------------
# Security
# ----------------------------------------------------------------
# IMPORTANT: Keep this secret secure!
JWT_SECRET=$JWT_SECRET

# ================================================================
# DO NOT COMMIT THIS FILE TO VERSION CONTROL
# ================================================================
EOF

chmod 600 "$ENV_FILE"
log_info "Configuration saved to $ENV_FILE (permissions: 600)"

# ----------------------------------------------------------------
# Summary
# ----------------------------------------------------------------

echo ""
echo "================================================================"
log_info "Setup Complete! 🎉"
echo "================================================================"
echo ""
echo "Configuration Summary:"
echo "  📦 Docker Images: ${DOCKER_USERNAME}/filelocker-*:latest"
echo "  🌐 Web Port: $WEB_PORT"
echo "  🗄️  Database: filelocker_db"
echo "  📁 Object Storage: MinIO (filelocker-files)"
echo "  🔒 Secrets: Generated securely"
echo ""
echo "Next Steps:"
echo ""
echo "  1. Pull Docker images:"
echo "     $DOCKER_CMD -f docker-compose.yml pull"
echo ""
echo "  2. Start File Locker:"
echo "     $DOCKER_CMD -f docker-compose.yml up -d"
echo ""
echo "  3. Check status:"
echo "     $DOCKER_CMD -f docker-compose.yml ps"
echo ""
echo "  4. View logs:"
echo "     $DOCKER_CMD -f docker-compose.yml logs -f"
echo ""
echo "  5. Access the application:"
echo "     🌐 Web: http://localhost:$WEB_PORT"
echo "     👤 Default Login: See config.yaml (security.default_admin)"
echo "     ⚠️  CHANGE THE ADMIN PASSWORD IN config.yaml BEFORE FIRST START!"
echo ""
echo "================================================================"
echo ""

# ----------------------------------------------------------------
# Offer to start services
# ----------------------------------------------------------------

read -p "Do you want to start File Locker now? (Y/n): " START_NOW

if [[ -z "$START_NOW" ]] || [[ "$START_NOW" =~ ^[Yy]$ ]]; then
    log_step "Pulling Docker images..."
    $DOCKER_CMD -f docker-compose.yml pull
    
    log_step "Starting services..."
    $DOCKER_CMD -f docker-compose.yml up -d
    
    echo ""
    log_info "Services started! Waiting for health checks..."
    sleep 5
    
    echo ""
    $DOCKER_CMD -f docker-compose.yml ps
    
    echo ""
    log_info "File Locker is starting up!"
    log_info "Access it at: http://localhost:$WEB_PORT"
else
    log_info "Skipping service start. Run the commands above when you're ready!"
fi
