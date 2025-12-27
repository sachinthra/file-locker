# File Locker Backend 🔐

Secure file storage backend with encryption, video streaming, and automatic cleanup features.

## 📋 Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Building & Deployment](#building--deployment)
- [Development](#development)
- [Troubleshooting](#troubleshooting)

## ✨ Features

- 🔒 **End-to-End Encryption**: AES-256-CTR encryption for all files
- 🎬 **Video Streaming**: HTTP range request support for video playback
- 🔑 **JWT Authentication**: Secure token-based authentication
- ⏰ **Auto Cleanup**: Configurable automatic deletion of expired files
- 📦 **Object Storage**: MinIO/S3-compatible storage backend
- ⚡ **Redis Caching**: Fast metadata access and session management
- 🔄 **gRPC API**: High-performance RPC for file metadata operations
- 🚦 **Rate Limiting**: Per-user request throttling
- 📊 **Health Checks**: Built-in health monitoring endpoints

## 📦 Prerequisites

- **Go**: 1.21 or higher
- **Docker & Docker Compose**: For running MinIO and Redis
- **Make**: For build automation (optional)
- **Protocol Buffers**: For gRPC code generation (optional)

## 🚀 Quick Start

### 1. Clone and Setup

```bash
cd backend
cp configs/config.yaml.example configs/config.yaml  # If example exists
```

### 2. Start Dependencies

```bash
# From project root
docker-compose up -d minio redis

# Or just MinIO and Redis
docker run -d -p 9000:9000 -p 9001:9001 \
  --name minio \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"

docker run -d -p 6379:6379 \
  --name redis \
  redis:alpine
```

### 3. Install Dependencies

```bash
go mod download
go mod vendor  # Optional: vendor dependencies
```

### 4. Run the Server

```bash
go run cmd/server/main.go
```

Server will start on:
- **HTTP API**: http://localhost:9010
- **gRPC API**: localhost:9011

## ⚙️ Configuration

Edit `configs/config.yaml`:

```yaml
server:
  port: 9010
  grpc_port: 9011
  host: "0.0.0.0"
  read_timeout: 60s
  write_timeout: 60s

security:
  jwt_secret: "your-secret-key-change-this"
  session_timeout: 3600  # seconds

storage:
  minio:
    endpoint: "localhost:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    bucket: "filelocker"
    use_ssl: false
    region: "us-east-1"
  
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0

features:
  auto_delete:
    enabled: true
    check_interval: 60  # minutes
  
  video_streaming:
    enabled: true
    chunk_size: 1048576  # 1MB
```

## 📚 API Documentation

### Authentication

#### Register User
```bash
curl -X POST http://localhost:9010/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "securepass123",
    "email": "test@example.com"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "uuid-here",
  "expires_at": "2025-12-28T10:00:00Z"
}
```

#### Login
```bash
curl -X POST http://localhost:9010/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "securepass123"
  }'
```

#### Logout
```bash
curl -X POST http://localhost:9010/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### File Operations

#### Upload File
```bash
curl -X POST http://localhost:9010/api/v1/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/file.mp4" \
  -F "tags=video,demo" \
  -F "expire_after=86400"  # Optional: seconds until expiry
```

**Response:**
```json
{
  "file_id": "uuid-here",
  "file_name": "file.mp4",
  "size": 1048576,
  "mime_type": "video/mp4",
  "created_at": "2025-12-27T15:30:00Z",
  "expires_at": "2025-12-28T15:30:00Z"
}
```

#### List Files
```bash
curl http://localhost:9010/api/v1/files \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Search Files
```bash
curl "http://localhost:9010/api/v1/files/search?q=video" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Download File
```bash
curl http://localhost:9010/api/v1/download/FILE_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o downloaded_file.mp4
```

#### Stream Video (Range Support)
```bash
# Full stream
curl http://localhost:9010/api/v1/stream/FILE_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o video.mp4

# Partial stream (seeking)
curl http://localhost:9010/api/v1/stream/FILE_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Range: bytes=0-1048576" \
  -o video_chunk.mp4
```

#### Delete File
```bash
curl -X DELETE "http://localhost:9010/api/v1/files?id=FILE_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Health Check
```bash
curl http://localhost:9010/health
```

## 🧪 Testing

### Run All Tests
```bash
make test
# Or
go test ./... -v -cover
```

### Run Specific Package Tests
```bash
# Test storage layer
go test ./internal/storage/... -v

# Test API handlers
go test ./internal/api/... -v

# Test with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Manual Testing Workflow

#### 1. Setup Test Environment
```bash
# Export test variables
export API_BASE="http://localhost:9010/api/v1"
export TEST_FILE="test/test_video.mp4"
```

#### 2. Register & Login
```bash
# Register
curl -X POST $API_BASE/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123","email":"test@test.com"}' \
  | jq -r '.token' > /tmp/token.txt

# Save token
export TOKEN=$(cat /tmp/token.txt)
```

#### 3. Upload Test File
```bash
curl -X POST $API_BASE/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$TEST_FILE" \
  | jq '.'
```

#### 4. List & Search Files
```bash
# List all files
curl $API_BASE/files \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# Search for specific file
curl "$API_BASE/files/search?q=test" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# Save first file ID
export FILE_ID=$(curl -s $API_BASE/files \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.[0].file_id')
```

#### 5. Download & Stream
```bash
# Download file
curl $API_BASE/download/$FILE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -o downloaded.mp4

# Stream with range request
curl $API_BASE/stream/$FILE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Range: bytes=0-1000000" \
  -o streamed_chunk.mp4

# Verify downloaded file
ls -lh downloaded.mp4
md5sum downloaded.mp4
```

#### 6. Browser Stream Test
```bash
# Open test page
open test/stream-test.html

# Or use this simple test
echo "
<video controls width='800'>
  <source id='videoSrc' type='video/mp4'>
</video>
<script>
  fetch('$API_BASE/stream/$FILE_ID', {
    headers: {'Authorization': 'Bearer $TOKEN'}
  })
  .then(r => r.blob())
  .then(blob => {
    document.getElementById('videoSrc').src = URL.createObjectURL(blob);
  });
</script>
" > /tmp/test-player.html && open /tmp/test-player.html
```

### Integration Tests

```bash
# Run with test database
TEST_MODE=true go test ./... -v

# Run with race detector
go test ./... -race

# Benchmark tests
go test ./... -bench=. -benchmem
```

## 🏗️ Building & Deployment

### Build Binary

```bash
# Using Makefile
make build

# Or manually
go build -o filelocker cmd/server/main.go

# Run binary
./filelocker
```

### Build with Version Info

```bash
VERSION=1.0.0
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse HEAD)

go build -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
  -o filelocker cmd/server/main.go
```

### Docker Build

```bash
# Build image
make docker-build
# Or
docker build -t filelocker-backend:latest .

# Run container
docker run -d \
  --name filelocker \
  -p 9010:9010 \
  -p 9011:9011 \
  -v $(pwd)/configs:/root/configs \
  --link minio:minio \
  --link redis:redis \
  filelocker-backend:latest
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

## 🛠️ Development

### Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/                     # HTTP handlers
│   │   ├── auth.go
│   │   ├── upload.go
│   │   ├── download.go
│   │   ├── stream.go
│   │   ├── files.go
│   │   └── utils.go
│   ├── auth/                    # Authentication
│   │   ├── jwt.go
│   │   └── middleware.go
│   ├── config/                  # Configuration
│   │   └── config.go
│   ├── crypto/                  # Encryption
│   │   └── aes.go
│   ├── grpc/                    # gRPC services
│   │   └── file_service.go
│   ├── storage/                 # Storage layer
│   │   ├── minio.go
│   │   └── redis.go
│   └── worker/                  # Background workers
│       └── cleanup.go
├── pkg/
│   └── proto/                   # Protocol buffers
│       ├── file_service.proto
│       ├── file_service.pb.go
│       └── file_service_grpc.pb.go
├── configs/
│   └── config.yaml              # Configuration file
├── test/                        # Test files and utilities
│   ├── stream-test.html
│   └── test_video.mp4
├── Dockerfile
├── Makefile
└── README.md
```

### Generate Protobuf Code

```bash
make proto
# Or
cd pkg/proto
protoc --go_out=. --go-grpc_out=. file_service.proto
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

### Add New Dependency

```bash
go get github.com/example/package
go mod tidy
go mod vendor  # Optional
```

## 🐛 Troubleshooting

### Server Won't Start

**Problem**: "Failed to connect to MinIO/Redis"

**Solution**:
```bash
# Check if services are running
docker ps | grep -E 'minio|redis'

# Restart services
docker-compose restart minio redis

# Check logs
docker-compose logs minio redis
```

### CORS Errors in Browser

**Problem**: "Access blocked by CORS policy"

**Solution**: Uncomment the correct CORS headers in `cmd/server/main.go`:
```go
AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:9010", "null"},
AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "Range"},
ExposedHeaders: []string{"Content-Length", "Content-Range", "Accept-Ranges", "Content-Type"},
```

### Video Won't Play

**Problem**: Video downloads but won't play

**Checklist**:
1. Check file is actually a video: `file downloaded.mp4`
2. Verify encryption/decryption: Compare original and downloaded file sizes
3. Check browser console for errors
4. Try downloading with curl first to isolate streaming issues

**Debug**:
```bash
# Compare file sizes (should be close, encrypted slightly larger)
ls -lh original.mp4 downloaded.mp4

# Check if file is corrupted
ffprobe downloaded.mp4
```

### Upload Fails

**Problem**: "Failed to upload file"

**Solution**:
```bash
# Check file size limit in config
# Check MinIO storage space
mc admin info myminio

# Check Redis memory
redis-cli info memory

# Verify file permissions
ls -la test/test_video.mp4
```

### Token Expired

**Problem**: 401 Unauthorized after some time

**Solution**:
```bash
# Login again to get new token
curl -X POST $API_BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123"}' \
  | jq -r '.token'

# Increase session timeout in config.yaml
security:
  session_timeout: 7200  # 2 hours
```

### High Memory Usage

**Problem**: Server consuming too much memory

**Solutions**:
1. Reduce chunk sizes in config
2. Enable auto-cleanup for old files
3. Monitor with: `docker stats filelocker`
4. Adjust Redis maxmemory settings

### Port Already in Use

**Problem**: "bind: address already in use"

**Solution**:
```bash
# Find process using port
lsof -i :9010
lsof -i :9011

# Kill process
kill -9 <PID>

# Or change port in config.yaml
```

## 📝 Environment Variables

Override config with environment variables:

```bash
export SERVER_PORT=8080
export SERVER_GRPCPORT=8081
export SECURITY_JWTSECRET=your-secret-key
export STORAGE_MINIO_ENDPOINT=minio:9000
export STORAGE_MINIO_ACCESSKEY=admin
export STORAGE_MINIO_SECRETKEY=password
export STORAGE_REDIS_ADDR=redis:6379
```

## 🤝 Contributing

1. Run tests before committing
2. Follow Go conventions
3. Update tests for new features
4. Update this README if adding new endpoints

## 📄 License

See LICENSE file in project root.

## 🔗 Related Documentation

- [Implementation Guide](IMPLEMENTATION_GUIDE.md)
- [Architecture Docs](../Docs/ARCHITECTURE.md)
- [API Design](../Docs/Design.md)

---

**Need Help?** Check the troubleshooting section or create an issue with logs and steps to reproduce.
