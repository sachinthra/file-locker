# File Locker - Design Document

A secure file encryption utility. File Locker provides both command-line and graphical interfaces for managing encrypted files.

## Core Features
- **Server-Side Encryption:** Files are encrypted by the server before storage.
- **User Authentication:** Secure login protects access to keys.
- **Ease of Use:** Drag-and-drop web interface and simple CLI.
- **Standard Security:** Uses AES-256-GCM encryption managed by the backend.
- **Media Streaming:** Securely stream encrypted video/audio directly to the browser without downloading the full file.

## Technology Stack
- **Backend:** Go (Handles encryption, storage, and API)
- **Frontend:** Preact (Web Interface)
- **Communication:** HTTP/REST (File Uploads/Streaming) + gRPC (Control/Metadata)
- **Containerization:** Docker & Docker Compose
- **Storage:** MinIO (Object Storage)
- **Database:** Redis (Session/Metadata)

## Port Configuration

| Service | Port | Purpose |
|---------|------|---------|
| HTTP Server | 9010 | Web UI, REST API |
| gRPC Server | 9011 | Metadata operations |
| MinIO API | 9012 | Object storage API |
| MinIO Console | 9013 | MinIO dashboard |
| Redis | 6379 | Cache (default) |

## CLI Command Reference

```bash
# Authentication
filelocker login                    # Login to server
filelocker logout                   # Logout

# File Operations
filelocker upload <file>            # Upload and encrypt file
filelocker upload <file> --expire 1h  # Upload with expiration
filelocker list                     # List all files
filelocker search <query>           # Search files
filelocker download <file-id>       # Download and decrypt
filelocker delete <file-id>         # Delete file
filelocker info <file-id>           # Show file metadata

# Streaming
filelocker stream <file-id>         # Stream video/audio

# Batch Operations
filelocker upload file1 file2 file3 # Upload multiple files

# Configuration
filelocker config set server.url http://localhost:9010
filelocker config show              # Show current configuration
```

## Implementation Roadmap

### Phase 1: Core Backend
- [ ] Implement Server-Side AES-256-GCM encryption service.
- [ ] Setup MinIO for storing encrypted blobs.
- [ ] Create HTTP API for streaming file uploads (`POST /upload`).
- [ ] Implement User Authentication (JWT).

### Phase 2: Web Frontend
- [ ] Build Preact File Manager UI.
- [ ] **Drag-and-Drop:** Implement Dropzone area for uploads.
- [ ] **Progress Indicators:** Visual bars for upload/download status using Axios/XHR.
- [ ] **Video Streaming:** Serve decrypted video streams directly to browser player using HTTP Range headers.

### Phase 3: CLI & Advanced
- [ ] Build CLI tool using Cobra.
- [ ] Implement Batch Operations (upload multiple files via parallel requests).
- [ ] Add "Auto-Delete" policy (Server background worker deletes file after set time).

## Security Model
- **Encryption at Rest:** Files are encrypted on the disk (MinIO) using AES-256-GCM.
- **Transport Security:** All data in transit is protected by TLS 1.3.
- **Access Control:** Only authenticated users can trigger decryption/download.
- **Key Management:** Keys are managed by the server and associated with the user's account.