# File Locker

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED?logo=docker)](https://docker.com)

**A secure, self-hosted file encryption and streaming utility.**

</div>

---

## Overview

File Locker is a robust file encryption server. It encrypts your files on the server-side before storing them, ensuring your data is safe at rest. It features a modern web interface with drag-and-drop support, video streaming capabilities, and a CLI for advanced users.

## Features

- 🔐 **Server-Side Encryption:** AES-256-GCM encryption handles everything automatically.
- 🎬 **Secure Streaming:** Watch encrypted videos directly in the browser without downloading.
- 🚀 **High Performance:** Go backend with MinIO storage for speed and scalability.
- 📤 **Easy Uploads:** Drag-and-drop interface with batch processing and progress bars.
- ⏲️ **Auto-Delete:** Set files to automatically expire and vanish after a set time.
- 📦 **Docker Ready:** Deploy in minutes with Docker Compose.

## Quick Start

### Method 1: Docker Compose (Recommended)

```bash
git clone [https://github.com/](https://github.com/)[username]/file-locker.git
cd file-locker

# Start all services (App, MinIO, Redis)
docker-compose up -d

# Access the web interface
# Open your browser to http://localhost:9010

```

### Method 2: Build from Source

```bash
# Clone repository
git clone https://github.com/[username]/file-locker.git
cd file-locker

# Build backend
cd backend
go build -o filelocker cmd/server/main.go

# Build frontend
cd ../frontend
npm install
npm run build

# Run with docker-compose
cd ..
docker-compose up -d
```

## CLI Usage

```bash
# Upload a file (encrypts and stores on server)
filelocker upload my-video.mp4
# or with short alias
fl upload my-video.mp4

# Upload with auto-delete (expires after 1 hour)
filelocker upload document.pdf --expire 1h

# Upload multiple files (batch operation)
filelocker upload file1.jpg file2.png file3.pdf

# List all your files
filelocker list

# Search for files
filelocker search "vacation"

# Download and decrypt a file
filelocker download <file-id>

# Get file info without downloading
filelocker info <file-id>

# Delete a file
filelocker delete <file-id>

# Stream video directly (opens in browser)
filelocker stream <file-id>
```

## Architecture

File Locker uses a standard **Server-Side Encryption** architecture:

* **Frontend:** Preact app communicating via REST (Files) and gRPC (Metadata).
* **Backend:** Go server handling encryption streams.
* **Storage:** MinIO stores the encrypted blobs.

See [ARCHITECTURE.md](ARCHITECTURE.md) for technical details and [IMPLEMENTATION.md](IMPLEMENTATION.md) for developer guides.

## Configuration

### Environment Variables

```bash
# Server Configuration
SERVER_PORT=9010              # HTTP/REST API port
GRPC_PORT=9011                # gRPC metadata service port

# MinIO Configuration
MINIO_ENDPOINT=localhost:9012  # MinIO API endpoint
MINIO_CONSOLE=localhost:9013   # MinIO web console
MINIO_ACCESS_KEY=minioadmin    # MinIO access key
MINIO_SECRET_KEY=minioadmin    # MinIO secret key
MINIO_BUCKET=filelocker        # Bucket name for encrypted files

# Redis Configuration
REDIS_ADDR=localhost:6379      # Redis address
REDIS_PASSWORD=                # Redis password (optional)
REDIS_DB=0                     # Redis database number

# Security
JWT_SECRET=your-secret-key-here  # Change in production!
SESSION_TIMEOUT=3600             # Session timeout in seconds
AUTO_DELETE_ENABLED=true         # Enable auto-delete feature
```

### Port Reference

| Service | Port | Purpose |
|---------|------|----------|
| HTTP Server | 9010 | Web UI, File uploads/downloads |
| gRPC Server | 9011 | Metadata API, control operations |
| MinIO API | 9012 | Object storage API |
| MinIO Console | 9013 | MinIO web dashboard |
| Redis | 6379 | Session/metadata cache |

## Roadmap

See [Design.md](Design.md) for the detailed implementation roadmap.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">
Made with 🔐 by [Your Name]
</div>