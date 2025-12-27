# Contributing to File Locker

Thank you for your interest in contributing to File Locker!

## Getting Started

### Prerequisites
- **Go:** 1.21 or higher
- **Node.js:** 18 or higher
- **Docker:** 20.10 or higher
- **MinIO:** (Running via Docker)

### First Time Setup
```bash
git clone https://github.com/YOUR_USERNAME/file-locker.git
cd file-locker

# Run setup script
make dev-setup
# or
./scripts/setup-dev.sh
```

## Development Setup

### Backend Development

The backend runs a Hybrid server (gRPC on 9011, HTTP on 9010).

```bash
cd backend
go mod download
# Starts both gRPC and HTTP services
go run cmd/server/main.go

```

### Frontend Development (Web)

```bash
cd frontend
npm install
npm run dev

```

## Coding Standards

### Security Guidelines (Server-Side)

1. **TLS Enforcement:**
* Never disable TLS in production.
* All handlers must check for authentication before processing files.


2. **Memory Safety:**
* Use `defer` to close file streams and database connections.
* When decrypting, do not buffer the entire file in RAM; always stream.


3. **Input Validation:**
* Validate `Content-Type` and file extensions.
* Sanitize filenames to prevent path traversal attacks.



## Testing

* **Unit Tests:** `go test ./...` and `npm test`
* **Integration:** Ensure MinIO is running (`docker-compose up -d minio`) before running integration tests.

Thank you for contributing! 🔐
*Last Updated: December 26, 2025*
