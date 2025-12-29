# 📦 File Locker - Production Distribution Files

This directory contains everything needed for production deployment and distribution of File Locker.

---

## 📁 Directory Structure

```
install/
├── docker-compose.yml    Production compose (pre-built images)
└── setup.sh              Interactive setup script

backend/migrations/       SQL migration files (embedded in binary)
backend/internal/db/      Migration logic

scripts/
└── build-release.sh      Build all artifacts (Docker + CLI + .deb)

Docs/
├── PRODUCTION_DISTRIBUTION.md    Comprehensive deployment guide
└── RASPBERRY_PI_DEPLOYMENT.md    Raspberry Pi specific guide

PRODUCTION_QUICK_START.md         Quick reference cheat sheet
IMPLEMENTATION_COMPLETE.md         Implementation summary
```

---

## 🚀 Quick Start

### For Developers (Build Everything)

```bash
export DOCKER_USERNAME="yourusername"
./scripts/build-release.sh
```

**Generates:**
- Docker images → Docker Hub
- CLI binaries → `bin/fl-*`
- Debian structure → `dist/deb/`
- Release archives → `dist/*.tar.gz`

---

### For Server Admins (Deploy)

**Method 1: Interactive Setup**
```bash
cd install/
./setup.sh
```

**Method 2: Manual**
```bash
# Create .env file (see example below)
docker compose -f install/docker-compose.yml pull
docker compose -f install/docker-compose.yml up -d
```

**Example `.env`:**
```bash
DOCKER_USERNAME=yourusername
WEB_PORT=8080
DB_USER=filelocker_user
DB_PASSWORD=<random-32-chars>
DB_NAME=filelocker_db
MINIO_ACCESS_KEY=filelocker_minio
MINIO_SECRET_KEY=<random-32-chars>
MINIO_BUCKET=filelocker-files
JWT_SECRET=<random-32-chars>
```

Generate passwords: `openssl rand -base64 32 | tr -d "=+/" | cut -c1-32`

---

### For End Users (CLI)

**Install:**
```bash
# Mac (ARM)
curl -LO <release-url>/fl-darwin-arm64
chmod +x fl-darwin-arm64 && sudo mv fl-darwin-arm64 /usr/local/bin/fl

# Mac (Intel)
curl -LO <release-url>/fl-darwin-amd64
chmod +x fl-darwin-amd64 && sudo mv fl-darwin-amd64 /usr/local/bin/fl

# Windows
# Download fl-windows-amd64.exe, rename to fl.exe, add to PATH

# Linux
curl -LO <release-url>/fl-linux-amd64
chmod +x fl-linux-amd64 && sudo mv fl-linux-amd64 /usr/local/bin/fl
```

**Use:**
```bash
fl login --host http://your-server:8080 -u admin -p password123
fl upload document.pdf
fl ls
fl download <file-id> -o downloaded.pdf
```

---

## ✅ What's Included

### Auto-Migrations ✅
- SQL files embedded in Go binary
- Runs automatically on server startup
- No external SQL files needed
- Fail-fast if migrations fail

### Smart CLI ✅
- `--host` flag with auto-path detection
- Saves configuration to `~/.filelocker/config.json`
- Supports both username/password and PAT login

### Production Docker Compose ✅
- Pre-built images (no source code needed)
- Only frontend port exposed (user-selectable)
- Backend accessible only via Nginx proxy
- Optimized for Raspberry Pi

### Interactive Setup ✅
- Auto-detects Docker version
- Smart port selection
- Generates secure random passwords
- Creates `.env` file

### Build System ✅
- Multi-platform Docker images (amd64 + arm64)
- CLI binaries for Mac/Windows/Linux
- Debian package structure
- Automated release process

---

## 📚 Documentation

- **[PRODUCTION_QUICK_START.md](./PRODUCTION_QUICK_START.md)** - Quick reference
- **[Docs/PRODUCTION_DISTRIBUTION.md](./Docs/PRODUCTION_DISTRIBUTION.md)** - Complete guide
- **[Docs/RASPBERRY_PI_DEPLOYMENT.md](./Docs/RASPBERRY_PI_DEPLOYMENT.md)** - Raspberry Pi guide
- **[IMPLEMENTATION_COMPLETE.md](./IMPLEMENTATION_COMPLETE.md)** - Implementation summary

---

## 🏗️ Architecture

```
CLI Client (Mac/Windows/Linux)
    ↓ HTTP/HTTPS
    ↓
Frontend Container (Port: 8080)
    ├── Serves Preact SPA
    └── Proxies /api → Backend
        ↓
Backend Container (Internal Only)
    ├── Auto-runs migrations on startup
    ├── REST API: /api/v1/*
    └── gRPC API: /grpc/*
        ↓
PostgreSQL + MinIO + Redis (Internal Only)
```

---

## 🔒 Security Features

- ✅ Only frontend port exposed
- ✅ Backend not accessible from host
- ✅ Database services internal only
- ✅ Auto-generated secure passwords
- ✅ JWT-based authentication
- ✅ Role-based access control (RBAC)

---

## 📊 Resource Usage

| Service | RAM | CPU | Disk |
|---------|-----|-----|------|
| Frontend | ~10MB | <1% | ~50MB |
| Backend | ~30MB | 1-5% | ~50MB |
| PostgreSQL | ~20MB | 1-3% | ~100MB |
| MinIO | ~40MB | 1-5% | Variable |
| Redis | ~10MB | <1% | ~20MB |
| **TOTAL** | **~110MB** | **<15%** | **~220MB + data** |

*Tested on Raspberry Pi 4 (4GB RAM)*

---

## 🎯 Supported Platforms

### Server
- ✅ Raspberry Pi 4 (ARM64)
- ✅ Ubuntu/Debian (ARM64, AMD64)
- ✅ macOS (Docker Desktop)
- ✅ Any Linux with Docker

### CLI
- ✅ Mac (M1/M2/M3 ARM64)
- ✅ Mac (Intel AMD64)
- ✅ Windows (AMD64)
- ✅ Linux (AMD64)
- ✅ Linux (ARM64)

---

## 🐛 Troubleshooting

### Build Issues
```bash
# Docker login required
docker login

# Update Docker Desktop for buildx support
# Ensure Go 1.21+ installed
```

### Server Won't Start
```bash
# Check migration logs
docker logs filelocker-server | grep migration

# Check all services
docker compose -f install/docker-compose.yml ps

# View logs
docker compose -f install/docker-compose.yml logs -f
```

### CLI Can't Connect
```bash
# Test server accessibility
curl http://your-server:8080/health

# Re-login with correct host
fl login --host http://your-server:8080 -u admin -p password123

# Check saved config
cat ~/.filelocker/config.json
```

---

## 📝 Default Credentials

**Username:** `admin`  
**Password:** `password123`  
⚠️ **CHANGE IMMEDIATELY AFTER FIRST LOGIN!**

---

## 🤝 Contributing

See [Docs/CONTRIBUTING.md](./Docs/CONTRIBUTING.md)

---

## 📄 License

See [LICENSE](./LICENSE)

---

## 🎉 Ready for Production!

All components are tested and ready for production deployment:
- ✅ Multi-platform Docker images
- ✅ Native CLI binaries
- ✅ Automated migrations
- ✅ Interactive installer
- ✅ Comprehensive documentation
- ✅ Minimal resource usage
- ✅ Secure by default

**Get Started:** Read [PRODUCTION_QUICK_START.md](./PRODUCTION_QUICK_START.md)
