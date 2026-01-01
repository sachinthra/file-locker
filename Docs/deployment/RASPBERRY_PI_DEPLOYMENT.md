# 🚀 Raspberry Pi Production Deployment Guide

## Overview

This guide explains how to deploy File Locker to a Raspberry Pi 4 (ARM64) using **pre-built Docker images**. No source code needed on the Pi.

---

## 📋 Prerequisites

### On Development Machine (x86/amd64)
- Docker Desktop with buildx support
- Docker Hub account
- Git repository with File Locker source code

### On Raspberry Pi 4
- Raspberry Pi OS (64-bit) or Ubuntu Server ARM64
- Docker installed (`curl -fsSL https://get.docker.com | sh`)
- Docker Compose V2 installed
- At least 4GB RAM recommended
- 16GB+ SD card or SSD

---

## 🔨 Step 1: Build ARM64 Images (Development Machine)

### 1.1 Update Docker Hub Username

Edit `docker-compose.prod.yml` and replace `youruser` with your Docker Hub username:

```yaml
services:
  frontend:
    image: youruser/filelocker-frontend:latest  # Change 'youruser'
  filelocker:
    image: youruser/filelocker-backend:latest   # Change 'youruser'
```

### 1.2 Build and Push Images

```bash
# Set your Docker Hub username
export DOCKER_USERNAME="your-docker-username"

# Run the build script
./scripts/build-and-push-arm64.sh
```

This will:
- Build both images for ARM64 architecture
- Tag with `:latest` and `:YYYYMMDD` date stamps
- Push to Docker Hub

**Expected output:**
```
✅ Backend image built and pushed successfully!
   Image: youruser/filelocker-backend:latest
✅ Frontend image built and pushed successfully!
   Image: youruser/filelocker-frontend:latest
```

---

## 📦 Step 2: Prepare Raspberry Pi Files

### 2.1 Create Deployment Directory on Pi

```bash
# SSH into your Raspberry Pi
ssh pi@raspberrypi.local

# Create deployment directory
mkdir -p ~/filelocker-deploy
cd ~/filelocker-deploy
```

### 2.2 Copy Required Files from Dev Machine

**Option A: Using SCP**

```bash
# On development machine
scp docker-compose.prod.yml pi@raspberrypi.local:~/filelocker-deploy/
scp .env pi@raspberrypi.local:~/filelocker-deploy/
scp -r configs pi@raspberrypi.local:~/filelocker-deploy/
scp -r backend/init-db pi@raspberrypi.local:~/filelocker-deploy/backend/
```

**Option B: Using Git (Recommended)**

```bash
# On Raspberry Pi
cd ~/filelocker-deploy
git clone https://github.com/yourusername/filelocker.git .

# Keep only necessary files
rm -rf backend/cmd backend/internal backend/pkg frontend/src
```

### 2.3 Verify File Structure

```
~/filelocker-deploy/
├── docker-compose.prod.yml
├── .env
├── configs/
│   └── config.yaml
└── backend/
    └── init-db/
        ├── 001_init.sql
        └── 002_add_user_roles.sql
```

---

## ⚙️ Step 3: Configure Environment Variables

### 3.1 Edit `.env` File

```bash
nano .env
```

**Critical Variables:**
```bash
# Server Ports
SERVER_PORT=9010
GRPC_PORT=9011

# Database
DB_USER=filelocker_user
DB_PASSWORD=change_this_secure_password
DB_NAME=filelocker_db
DB_PORT=5432

# MinIO Storage
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=change_this_secure_minio_password
MINIO_BUCKET=filelocker-files
MINIO_PORT_API=9000
MINIO_PORT_CONSOLE=9001

# Redis
REDIS_PORT=6379

# Security
JWT_SECRET=change_this_ultra_secure_jwt_secret_key_min_32_chars
```

### 3.2 Verify `config.yaml`

```bash
cat configs/config.yaml
```

Ensure logging paths are correct for Docker:
```yaml
logging:
  level: "info"  # Use 'info' in production
  path: "/var/log/filelocker/server.log"
  max_size_mb: 10
  max_backups: 3
  max_age_days: 28
```

---

## 🚀 Step 4: Deploy on Raspberry Pi

### 4.1 Pull Images

```bash
cd ~/filelocker-deploy
docker compose -f docker-compose.prod.yml pull
```

**Expected output:**
```
Pulling frontend    ... done
Pulling filelocker  ... done
Pulling postgres    ... done
Pulling minio       ... done
Pulling redis       ... done
```

### 4.2 Start Services

```bash
docker compose -f docker-compose.prod.yml up -d
```

### 4.3 Check Service Status

```bash
docker compose -f docker-compose.prod.yml ps
```

**Expected output:**
```
NAME                    STATUS        PORTS
filelocker-frontend     Up (healthy)  0.0.0.0:80->80/tcp
filelocker-server       Up (healthy)  
filelocker-postgres     Up (healthy)  0.0.0.0:5432->5432/tcp
filelocker-minio        Up (healthy)  0.0.0.0:9000-9001->9000-9001/tcp
filelocker-redis        Up (healthy)  0.0.0.0:6379->6379/tcp
```

### 4.4 View Logs

```bash
# All services
docker compose -f docker-compose.prod.yml logs -f

# Specific service
docker compose -f docker-compose.prod.yml logs -f filelocker

# Backend application logs (inside container)
docker exec filelocker-server cat /var/log/filelocker/server.log
```

---

## 🧪 Step 5: Verify Deployment

### 5.1 Access Web Interface

Open browser and navigate to:
```
http://raspberrypi.local
```
or
```
http://<raspberry-pi-ip-address>
```

### 5.2 Test Health Endpoints

```bash
# Frontend health
curl http://localhost/health

# Backend health (via Nginx proxy)
curl http://localhost/api/v1/health
```

### 5.3 Create Admin User

```bash
# Default admin user is created automatically during database initialization
# Username: admin
# Password: password123 (CHANGE THIS IMMEDIATELY)
```

To change admin password, connect to PostgreSQL:
```bash
docker exec -it filelocker-postgres psql -U filelocker_user -d filelocker_db
```

```sql
-- Generate new bcrypt hash (use online tool or backend API)
UPDATE users SET password = '$2a$10$new_bcrypt_hash_here' WHERE username = 'admin';
```

---

## 🔒 Step 6: Security Hardening (Production)

### 6.1 Remove External Port Exposure

Edit `docker-compose.prod.yml` and comment out these port mappings:

```yaml
postgres:
  # ports:
  #   - "${DB_PORT}:5432"  # Comment out for production

minio:
  # ports:
  #   - "${MINIO_PORT_API}:9000"
  #   - "${MINIO_PORT_CONSOLE}:9001"  # Comment out for production

redis:
  # ports:
  #   - "${REDIS_PORT}:6379"  # Comment out for production
```

Restart services:
```bash
docker compose -f docker-compose.prod.yml up -d
```

### 6.2 Enable Firewall

```bash
# Allow SSH and HTTP only
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw enable
```

### 6.3 Change Default Passwords

- Admin user password (see Step 5.3)
- Database password (`.env` file + restart)
- MinIO credentials (`.env` file + restart)
- JWT secret (`.env` file + restart)

---

## 📊 Monitoring & Maintenance

### Check Disk Usage

```bash
docker system df
```

### Clean Up Old Images

```bash
docker image prune -a
```

### View Container Resources

```bash
docker stats
```

### Backup Volumes

```bash
# Backup database
docker exec filelocker-postgres pg_dump -U filelocker_user filelocker_db > backup.sql

# Backup MinIO data
docker run --rm -v filelocker_minio-data:/data -v $(pwd):/backup alpine tar czf /backup/minio-backup.tar.gz /data
```

### Update Images

```bash
cd ~/filelocker-deploy
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

---

## 🐛 Troubleshooting

### Service Won't Start

```bash
# Check logs
docker compose -f docker-compose.prod.yml logs <service-name>

# Check if ports are already in use
sudo netstat -tulpn | grep -E ':(80|5432|6379|9000|9001)'
```

### Backend Can't Connect to Database

```bash
# Verify database is healthy
docker compose -f docker-compose.prod.yml ps postgres

# Check environment variables
docker compose -f docker-compose.prod.yml config
```

### Frontend Shows 502 Bad Gateway

```bash
# Check if backend is healthy
docker compose -f docker-compose.prod.yml ps filelocker

# Check Nginx logs
docker compose -f docker-compose.prod.yml logs frontend
```

### Out of Memory

```bash
# Check RAM usage
free -h

# Reduce Redis memory
# Edit docker-compose.prod.yml:
redis:
  command: redis-server --appendonly yes --maxmemory 128mb --maxmemory-policy allkeys-lru
```

---

## 📝 Architecture Notes

### Network Flow

```
Client Browser
    ↓ (Port 80)
Nginx Frontend Container
    ↓ (Internal Docker Network)
    ├── /api/* → Backend Container (Port 9010)
    ├── /grpc/* → Backend Container (Port 9011)
    └── /* → Preact Static Files

Backend Container
    ↓ (Internal Docker Network)
    ├── PostgreSQL (Port 5432)
    ├── MinIO (Port 9000)
    └── Redis (Port 6379)
```

### Why No Backend Ports Exposed?

- **Security**: Backend only accessible via Nginx proxy
- **Single Entry Point**: All traffic goes through port 80
- **Simplified Firewall**: Only need to allow port 80

### How to Access Backend Directly?

For debugging/CLI access, temporarily expose backend port:

```yaml
# In docker-compose.prod.yml
filelocker:
  ports:
    - "9010:9010"  # Temporary for debugging
```

---

## 📚 Additional Resources

- [Docker on Raspberry Pi](https://docs.docker.com/engine/install/debian/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Nginx Reverse Proxy Guide](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/)

---

## 🎯 Summary Checklist

- [ ] Build ARM64 images on dev machine
- [ ] Push images to Docker Hub
- [ ] Copy deployment files to Raspberry Pi
- [ ] Configure `.env` with secure passwords
- [ ] Pull images on Raspberry Pi
- [ ] Start services with `docker compose up -d`
- [ ] Verify health checks pass
- [ ] Test web interface access
- [ ] Change default admin password
- [ ] Comment out unnecessary port exposures
- [ ] Enable firewall
- [ ] Set up automated backups

---

**🎉 Congratulations! Your File Locker is now running on Raspberry Pi!**
