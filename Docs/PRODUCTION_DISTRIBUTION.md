# 🚀 Production Distribution Guide

This guide covers building and distributing File Locker for production use with **native installers** (.deb) and **cross-platform CLI binaries**.

---

## 📋 Architecture Overview

### Distribution Strategy

- **Server**: Pre-built Docker images on Docker Hub (no source code on server)
- **Server Installer**: `.deb` package or tar.gz archive
- **Client CLI**: Native binaries for Mac, Windows, Linux

### Network Architecture

```
Client CLI (Mac/Windows/Linux)
    ↓ HTTP/HTTPS
    ↓
Nginx Frontend Container (Port: User-Selected, e.g., 8080)
    ↓ Internal Docker Network
    ↓
Backend Container (No External Ports)
    ↓
PostgreSQL + MinIO + Redis (Internal Only)
```

**Key Points:**
- Only frontend exposes a port (user-selectable during setup)
- Backend accessed via Nginx `/api` proxy
- CLI talks to frontend, not backend directly
- Database migrations run automatically on backend startup

---

## 🏗️ Building for Release

### Prerequisites

- Docker Desktop with buildx support
- Go 1.21+ installed
- Docker Hub account
- Linux/macOS (script not tested on Windows)

### Step 1: Build Everything

```bash
# Set your Docker Hub username
export DOCKER_USERNAME="yourusername"

# Run the comprehensive build script
./scripts/build-release.sh
```

This script will:
1. ✅ Build multi-platform Docker images (amd64 + arm64)
2. ✅ Push images to Docker Hub
3. ✅ Build CLI binaries for all platforms
4. ✅ Create Debian package structure
5. ✅ Generate release archives

### Build Output

```
bin/
├── fl-darwin-arm64       # Mac M1/M2/M3
├── fl-darwin-amd64       # Mac Intel
├── fl-linux-amd64        # Linux x86_64
├── fl-linux-arm64        # Raspberry Pi / ARM64
└── fl-windows-amd64.exe  # Windows

dist/
├── deb/
│   ├── DEBIAN/
│   │   └── control       # Debian package metadata
│   └── opt/
│       └── filelocker/
│           ├── docker-compose.yml
│           ├── setup.sh
│           ├── config.yaml
│           └── README.txt
├── filelocker-server-1.0.0.tar.gz
└── filelocker-cli-latest.tar.gz
```

### Step 2: Create .deb Package (Optional)

```bash
# Requires dpkg-deb (available on Linux/macOS with Homebrew)
dpkg-deb --build dist/deb dist/filelocker-1.0.0.deb
```

---

## 📦 Server Installation

### Method 1: Debian Package (.deb)

**On Debian/Ubuntu/Raspberry Pi OS:**

```bash
# Install the package
sudo dpkg -i filelocker-1.0.0.deb

# Navigate to installation directory
cd /opt/filelocker

# Run interactive setup
./setup.sh

# Follow prompts to configure:
# - Docker Hub username
# - Web port (default: 8080)
# - Auto-generated secure passwords
```

### Method 2: Manual Installation (Any Linux/macOS)

```bash
# Extract server files
tar -xzf filelocker-server-1.0.0.tar.gz
cd opt/filelocker

# Run setup script
./setup.sh

# Or manually:
# 1. Create .env file (see template below)
# 2. Copy config.yaml to current directory
# 3. docker compose -f docker-compose.yml pull
# 4. docker compose -f docker-compose.yml up -d
```

### Manual .env Template

```bash
DOCKER_USERNAME=yourusername
WEB_PORT=8080
DB_USER=filelocker_user
DB_PASSWORD=<generate-random-32-chars>
DB_NAME=filelocker_db
MINIO_ACCESS_KEY=filelocker_minio
MINIO_SECRET_KEY=<generate-random-32-chars>
MINIO_BUCKET=filelocker-files
JWT_SECRET=<generate-random-32-chars>
```

Generate passwords:
```bash
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
```

---

## 💻 Client CLI Installation

### Mac (ARM64 - M1/M2/M3)

```bash
# Download binary
curl -LO https://github.com/youruser/filelocker/releases/latest/download/fl-darwin-arm64

# Make executable
chmod +x fl-darwin-arm64

# Move to PATH
sudo mv fl-darwin-arm64 /usr/local/bin/fl

# Verify installation
fl --version
```

### Mac (Intel)

```bash
curl -LO https://github.com/youruser/filelocker/releases/latest/download/fl-darwin-amd64
chmod +x fl-darwin-amd64
sudo mv fl-darwin-amd64 /usr/local/bin/fl
```

### Windows

```powershell
# Download fl-windows-amd64.exe
# Rename to fl.exe
# Add to PATH or use full path
```

### Linux

```bash
curl -LO https://github.com/youruser/filelocker/releases/latest/download/fl-linux-amd64
chmod +x fl-linux-amd64
sudo mv fl-linux-amd64 /usr/local/bin/fl
```

---

## 🔐 CLI Usage

### First-Time Setup

```bash
# Login with username/password
fl login --host http://your-server:8080 -u admin -p password123

# Or login with Personal Access Token (recommended)
fl login --host http://your-server:8080 --token <your-pat>
```

**Important:** The `--host` flag accepts:
- `http://localhost:8080` → Auto-appends `/api/v1`
- `http://raspberrypi.local:8080` → Auto-appends `/api/v1`
- `https://files.example.com` → Auto-appends `/api/v1`

Subsequent commands don't need `--host` (saved in `~/.filelocker/config.json`).

### Common Commands

```bash
# Upload a file
fl upload document.pdf

# Upload with options
fl upload photo.jpg --tags "vacation,2024" --expire 7d --desc "Beach photo"

# List files
fl ls

# List files (JSON output)
fl ls --json

# Download a file
fl download <file-id> -o downloaded-file.pdf

# Stream a file
fl stream <file-id>

# Delete a file
fl rm <file-id>

# Logout
fl logout
```

---

## 🔧 Server Management

### Check Status

```bash
cd /opt/filelocker  # or your installation directory
docker compose -f docker-compose.yml ps
```

### View Logs

```bash
# All services
docker compose -f docker-compose.yml logs -f

# Specific service
docker compose -f docker-compose.yml logs -f filelocker

# Backend logs (inside container)
docker exec filelocker-server cat /var/log/filelocker/server.log
```

### Update to Latest Version

```bash
# Pull new images
docker compose -f docker-compose.yml pull

# Restart services
docker compose -f docker-compose.yml up -d

# Migrations run automatically on startup
```

### Stop/Start Services

```bash
# Stop
docker compose -f docker-compose.yml down

# Start
docker compose -f docker-compose.yml up -d

# Restart specific service
docker compose -f docker-compose.yml restart filelocker
```

### Backup Data

```bash
# Backup database
docker exec filelocker-postgres pg_dump -U filelocker_user filelocker_db > backup.sql

# Backup MinIO files
docker run --rm \
  -v filelocker_minio-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/minio-backup.tar.gz /data

# Backup Redis
docker exec filelocker-redis redis-cli BGSAVE
```

---

## 🎯 Production Checklist

### Security

- [ ] Change default admin password immediately
- [ ] Generate strong random passwords (32+ characters)
- [ ] Use HTTPS with Let's Encrypt or reverse proxy
- [ ] Restrict firewall to only allow web port
- [ ] Keep Docker images updated
- [ ] Enable log rotation (handled automatically)

### Monitoring

- [ ] Set up health check monitoring
- [ ] Configure log aggregation
- [ ] Monitor disk usage (PostgreSQL, MinIO, Redis volumes)
- [ ] Set up alerts for service failures

### Performance (Raspberry Pi)

- [ ] Limit Redis memory (256MB default)
- [ ] Monitor CPU/RAM usage
- [ ] Use SSD instead of SD card for better I/O
- [ ] Consider disabling Redis persistence for cache-only usage

---

## 📊 Resource Usage

**Typical Resource Consumption:**

| Service | RAM | CPU | Disk |
|---------|-----|-----|------|
| Frontend | ~10MB | <1% | ~50MB |
| Backend | ~30MB | 1-5% | ~50MB |
| PostgreSQL | ~20MB | 1-3% | ~100MB |
| MinIO | ~40MB | 1-5% | Variable |
| Redis | ~10MB | <1% | ~20MB |
| **TOTAL** | ~110MB | <15% | ~220MB + data |

*Tested on Raspberry Pi 4 (4GB RAM)*

---

## 🐛 Troubleshooting

### Backend Won't Start

```bash
# Check logs
docker logs filelocker-server

# Common issues:
# - Database migration failed → Check PostgreSQL connection
# - MinIO unreachable → Check MinIO health
# - Redis connection failed → Check Redis health
```

### CLI Can't Connect

```bash
# Verify server is accessible
curl http://your-server:8080/health

# Should return: healthy

# Check saved config
cat ~/.filelocker/config.json

# Re-login with correct host
fl login --host http://your-server:8080 -u admin -p password123
```

### Database Migration Failed

Migrations run automatically on backend startup. If they fail:

```bash
# Check backend logs
docker logs filelocker-server | grep migration

# Migrations are embedded in the binary
# If schema is corrupted, you may need to reset:
docker compose down -v  # WARNING: Deletes all data!
docker compose up -d
```

---

## 📚 Additional Resources

- [Architecture Documentation](../Docs/ARCHITECTURE.md)
- [API Documentation](../Docs/API.md)
- [Contributing Guide](../Docs/CONTRIBUTING.md)
- [Raspberry Pi Deployment](../Docs/RASPBERRY_PI_DEPLOYMENT.md)

---

## 🎉 Summary

**What You Get:**

✅ Multi-platform Docker images (amd64 + arm64)  
✅ Native CLI binaries (Mac, Windows, Linux)  
✅ Debian package for easy installation  
✅ Interactive setup script with secure defaults  
✅ Auto-running database migrations  
✅ Production-ready configuration  
✅ Minimal resource usage (~110MB RAM)  
✅ Single web port exposure (security)  
✅ CLI with smart host detection  

**Perfect for:**
- 🏠 Home lab deployments
- 🥧 Raspberry Pi installations
- 🖥️ VPS/cloud servers
- 🏢 Internal company file sharing
- 👥 Small team collaboration

---

**Questions or Issues?**  
Open an issue on GitHub: https://github.com/youruser/filelocker/issues
