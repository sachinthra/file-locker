# 🚀 Quick Deployment Reference

## Pre-Deployment (Dev Machine)

```bash
# 1. Set your Docker Hub username
export DOCKER_USERNAME="your-docker-username"

# 2. Update docker-compose.prod.yml
sed -i '' "s/youruser/$DOCKER_USERNAME/g" docker-compose.prod.yml

# 3. Build and push ARM64 images
./scripts/build-and-push-arm64.sh

# Wait for: ✅ Backend image built and pushed successfully!
#           ✅ Frontend image built and pushed successfully!
```

---

## Deployment (Raspberry Pi)

```bash
# 1. Create directory
mkdir -p ~/filelocker-deploy && cd ~/filelocker-deploy

# 2. Copy files (choose one method)
# Option A: SCP from dev machine
scp docker-compose.prod.yml pi@raspberrypi.local:~/filelocker-deploy/
scp .env pi@raspberrypi.local:~/filelocker-deploy/
scp -r configs backend/init-db pi@raspberrypi.local:~/filelocker-deploy/

# Option B: Git clone
git clone https://github.com/yourusername/filelocker.git .

# 3. Edit secrets
nano .env  # Change all passwords and JWT secret

# 4. Pull images
docker compose -f docker-compose.prod.yml pull

# 5. Start services
docker compose -f docker-compose.prod.yml up -d

# 6. Check status
docker compose -f docker-compose.prod.yml ps

# Expected: All services show "Up (healthy)"
```

---

## Post-Deployment

```bash
# Test web interface
curl http://localhost/health  # Should return "healthy"

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Login to web interface
# URL: http://raspberrypi.local or http://<pi-ip-address>
# Default: admin / password123 (CHANGE IMMEDIATELY)
```

---

## Common Commands

```bash
# Start services
docker compose -f docker-compose.prod.yml up -d

# Stop services
docker compose -f docker-compose.prod.yml down

# Restart specific service
docker compose -f docker-compose.prod.yml restart filelocker

# View logs
docker compose -f docker-compose.prod.yml logs -f <service>

# Update images
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d

# Backup database
docker exec filelocker-postgres pg_dump -U filelocker_user filelocker_db > backup.sql

# Clean up
docker system prune -a
```

---

## File Locations on Pi

```
~/filelocker-deploy/
├── docker-compose.prod.yml  # Main orchestration
├── .env                      # Secrets (KEEP SECURE)
├── configs/
│   └── config.yaml          # App configuration
└── backend/
    └── init-db/             # Database initialization
        ├── 001_init.sql
        └── 002_add_user_roles.sql
```

---

## URLs & Ports

| Service | URL | Notes |
|---------|-----|-------|
| **Web Interface** | `http://raspberrypi.local` | Main entry point |
| **API** | `http://raspberrypi.local/api/v1` | Proxied through Nginx |
| **MinIO Console** | `http://raspberrypi.local:9001` | Optional (comment out in prod) |
| **PostgreSQL** | `raspberrypi.local:5432` | Optional (comment out in prod) |

---

## Security Checklist

- [ ] Change default admin password
- [ ] Update all passwords in `.env`
- [ ] Generate new JWT secret (32+ chars)
- [ ] Comment out unnecessary port exposures
- [ ] Enable firewall: `sudo ufw allow 80/tcp && sudo ufw enable`
- [ ] Keep `.env` file secure (never commit to Git)

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Service unhealthy | Check logs: `docker compose logs <service>` |
| Port already in use | `sudo netstat -tulpn \| grep :<port>` |
| Out of memory | Reduce Redis: `--maxmemory 128mb` |
| 502 Bad Gateway | Backend not healthy: `docker ps` |
| Can't login | Check database initialized: `docker logs filelocker-postgres` |

---

## Resource Usage (Typical)

| Service | RAM | CPU | Disk |
|---------|-----|-----|------|
| Frontend | ~10MB | <1% | ~50MB |
| Backend | ~30MB | 1-5% | ~50MB |
| PostgreSQL | ~20MB | 1-3% | ~100MB |
| MinIO | ~40MB | 1-5% | Variable |
| Redis | ~10MB | <1% | ~20MB |
| **TOTAL** | ~110MB | <15% | ~220MB + data |

*Raspberry Pi 4 with 4GB RAM can easily handle this workload*

---

## 📚 Full Documentation

For detailed instructions, see: [Docs/RASPBERRY_PI_DEPLOYMENT.md](../Docs/RASPBERRY_PI_DEPLOYMENT.md)
