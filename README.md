# File Locker 🔐

<div align="center">

**A secure, self-hosted file encryption and streaming server.**

[Features](https://www.google.com/search?q=%23-features) • [Quick Start](https://www.google.com/search?q=%23-quick-start) • [Configuration](https://www.google.com/search?q=%23%25EF%25B8%258F-configuration) • [API Documentation](https://www.google.com/search?q=%23-api-documentation) • [Testing](https://www.google.com/search?q=%23-testing)

</div>

---

## 📋 Overview

File Locker is a robust file encryption server designed for privacy and performance. It automatically encrypts files upon upload using **AES-256-GCM** (or CTR for streaming) before storing them in **MinIO**. It features a modern web interface, secure video streaming capabilities, and a developer-friendly API.

## ✨ Features

* 🔐 **Server-Side Encryption:** AES-256 encryption handles everything automatically at rest.
* 🎬 **Secure Streaming:** Watch encrypted videos directly in the browser with seeking support (HTTP Range requests).
* 🚀 **High Performance:** Go backend with MinIO object storage and Redis caching.
* 🐳 **Docker Ready:** Deploy in minutes with a "Single Source of Truth" configuration.
* 📦 **Batch Operations:** Drag-and-drop interface for multiple file uploads.
* ⏱️ **Auto-Cleanup:** Configurable automatic deletion of expired files.

---

## 🚀 Quick Start

### Prerequisites

* **Docker** & **Docker Compose**
* **Make** (Optional, but recommended)

### 1. Clone & Setup

```bash
git clone https://github.com/[username]/file-locker.git
cd file-locker

# Run the setup script to check dependencies
make dev-setup

```

### 2. Start Services

This command syncs your config, builds images, and starts the stack.

```bash
make docker-up
```

Access the services:

* **Backend API:** [http://localhost:9010](http://localhost:9010)
* **MinIO Console:** [http://localhost:9013](http://localhost:9013) (User/Pass: `minioadmin`)
* **API Health:** [http://localhost:9010/health](http://localhost:9010/health)

### 3. Start Frontend (Development)

In a separate terminal:

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at [http://localhost:5173](http://localhost:5173)

### 4. Stop Services

```bash
make docker-down
```

---

## 🌐 Frontend

The File Locker includes a modern web interface built with **Preact** and **Vite**.

### Features

- 🎨 Clean, responsive UI
- 📤 Drag-and-drop file upload with progress tracking
- 🔍 File search by name or tags
- 🏷️ Tag-based file organization
- ⏰ Optional file expiration
- 📥 One-click downloads
- 🎥 In-browser video streaming
- 🔐 JWT-based authentication

### Development

```bash
cd frontend
npm install      # Install dependencies
npm run dev      # Start dev server with hot reload
npm run build    # Build for production
npm run preview  # Preview production build
```

See [frontend/README.md](frontend/README.md) for detailed frontend documentation.

---

## ⚙️ Configuration

We use a **Single Source of Truth** strategy. You only edit **one file**: [`configs/config.yaml`](https://www.google.com/search?q=configs/config.yaml).

### Standard Config (`configs/config.yaml`)

```yaml
server:
  port: 9010          # Web UI & API Port
  grpc_port: 9011     # Internal gRPC Port

storage:
  minio:
    endpoint: "localhost:9012" # Host-accessible endpoint
    port_api: 9012             # Docker maps 9012 -> 9000
    port_console: 9013         # Docker maps 9013 -> 9001

```

*Note: When you run `make docker-up`, a script automatically syncs these ports to Docker Compose.*

---

## 📚 API Documentation & Manual Testing

You can test the API using `curl`. Set these variables first:

```bash
# Setup
export API_URL="http://localhost:9010/api/v1"

```

### 1. Authentication

**Register:**

```bash
curl -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin", "password":"password123", "email":"admin@local.host"}'

```

**Login (Get Token):**

```bash
# Save token to variable
export TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin", "password":"password123"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

echo "Token: $TOKEN"

```

### 2. File Operations

**Upload File:**

```bash
# Create dummy file
echo "This is a secret test file" > secret.txt

# Upload
curl -X POST "$API_URL/upload" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@secret.txt" \
  -F "tags=test,demo"

```

**List Files:**

```bash
curl -s "$API_URL/files" -H "Authorization: Bearer $TOKEN" | json_pp

```

**Get File ID:**

```bash
# Extract the first file ID from the list
export FILE_ID=$(curl -s "$API_URL/files" -H "Authorization: Bearer $TOKEN" | grep -o '"file_id":"[^"]*' | head -1 | grep -o '[^"]*$')
echo "File ID: $FILE_ID"

```

**Download File:**

```bash
curl -O -J \
  -H "Authorization: Bearer $TOKEN" \
  "$API_URL/download/$FILE_ID"

```

### 3. Video Streaming Test

If you uploaded a video file, you can test streaming support (Range requests).

```bash
# Fetch the first 1MB only (Test Seeking)
curl -v \
  -H "Authorization: Bearer $TOKEN" \
  -H "Range: bytes=0-1048576" \
  "$API_URL/stream/$FILE_ID" > video_chunk.mp4

```

---

## 🛠️ Development Workflow

### Directory Structure

```
.
├── backend/            # Go Backend
│   ├── cmd/            # Entry points
│   └── internal/       # Business logic (API, Auth, Crypto)
├── frontend/           # Preact Frontend
├── configs/            # Configuration (SSOT)
├── scripts/            # Helper scripts
└── docker-compose.yml  # Container definition

```

### Useful Make Commands

| Command | Description |
| --- | --- |
| `make dev` | Starts Backend and Frontend locally (hot-reload) |
| `make build` | Compiles binaries for production |
| `make test` | Runs unit tests for Go and JS |
| `make lint` | Runs code quality checks |
| `make clean` | Removes build artifacts |

### Running Tests

To run integration tests (requires Docker services running):

```bash
make test-backend

```

---

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](https://www.google.com/search?q=CONTRIBUTING.md) for details on code style and the pull request process.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.

---

<div align="center">
Made with 🔐 by [Your Name]
</div>