# File Locker - Architecture

## Table of Contents
- [System Architecture](#system-architecture)
- [Data Flow](#data-flow)
- [Component Details](#component-details)
- [Security Considerations](#security-considerations)

## System Architecture

The system uses a **Server-Side Encryption** model. The Client sends plaintext data over a secure TLS connection. The Server encrypts the data in memory and stores the ciphertext in MinIO.


```

┌──────────────┐          ┌─────────────────────────┐          ┌─────────────┐
│              │  HTTPS   │      Go Server          │  Write   │             │
│    Client    ├─────────►│  1. Receive Stream      ├─────────►│    MinIO    │
│ (Web / CLI)  │ Plaintext│  2. Encrypt (AES-256)   │ Encrypted│             │
│              │          │  3. Stream to Storage   │          │             │
└──────────────┘          └─────────────────────────┘          └─────────────┘

```

## API Endpoints

### REST API (Port 9010)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/auth/login` | User login (returns JWT) | No |
| `POST` | `/api/v1/auth/register` | User registration | No |
| `POST` | `/api/v1/upload` | Upload and encrypt file | Yes |
| `GET` | `/api/v1/files` | List user's files | Yes |
| `GET` | `/api/v1/files/{id}` | Get file metadata | Yes |
| `GET` | `/api/v1/download/{id}` | Download decrypted file | Yes |
| `GET` | `/api/v1/stream/{id}` | Stream decrypted media | Yes |
| `DELETE` | `/api/v1/files/{id}` | Delete file | Yes |
| `GET` | `/api/v1/search?q={query}` | Search files by name/tags | Yes |

### gRPC API (Port 9011)

```protobuf
service FileService {
  rpc GetFileMetadata(FileRequest) returns (FileMetadata);
  rpc ListFiles(ListRequest) returns (FileList);
  rpc UpdateTags(UpdateTagsRequest) returns (FileMetadata);
  rpc SetExpiration(ExpirationRequest) returns (FileMetadata);
}

service AdminService {
  rpc GetStats(Empty) returns (SystemStats);
  rpc ListUsers(ListUsersRequest) returns (UserList);
}
```

## Data Structures

### Redis Schema

```
# User Sessions
session:{token}
  - user_id: string
  - expires_at: timestamp
  - TTL: SESSION_TIMEOUT seconds

# File Metadata
file:{file_id}
  - user_id: string
  - filename: string
  - mime_type: string
  - size_bytes: int
  - encrypted_size_bytes: int
  - minio_path: string
  - encryption_key_id: string
  - created_at: timestamp
  - expires_at: timestamp (optional)
  - tags: []string
  - download_count: int

# User File Index
user:{user_id}:files
  - Set of file_ids

# Search Index
search:filename:{term}
  - Set of file_ids matching term
```

### MinIO Structure

```
Bucket: filelocker
├── {user_id}/
│   ├── {file_id}.encrypted       # Encrypted file data
│   └── {file_id}.metadata.json   # Backup metadata
```

## Data Flow

### Upload (Encryption)
1. **User** drags file to Web UI.
2. **Client** uploads file via HTTP `POST /api/v1/upload` (using Multipart or Binary stream).
3. **Server** authenticates the request via JWT.
4. **Server** generates a unique encryption key for the file.
5. **Server** streams the upload through an AES-256-GCM encrypter.
6. **Server** saves the *Encrypted* stream to MinIO at `{user_id}/{file_id}.encrypted`.
7. **Server** saves metadata (Filename, Key ID, Size) to Redis with key `file:{file_id}`.

### Download / Streaming (Decryption)
1. **User** requests file `GET /api/v1/download/{id}` or `<video src="/api/v1/stream/{id}">`.
2. **Server** authenticates user and checks permissions.
3. **Server** retrieves file metadata from Redis (`file:{file_id}`).
4. **Server** retrieves encrypted stream from MinIO.
5. **Server** decrypts the stream on-the-fly using stored encryption key.
6. **Client** receives plaintext stream.
   - *Note:* For videos, the server supports HTTP `Range` requests to allow seeking.
7. **Server** increments `download_count` in Redis.

## Component Details

```
backend/
├── cmd/
│   ├── server/              # Main server entry point
│   │   └── main.go
│   └── cli/                 # CLI tool
│       └── main.go

```
frontend/
├── src/
│   ├── components/
│   │   ├── FileUpload/       # Drag-and-drop upload
│   │   │   ├── index.jsx
│   │   │   └── styles.css
│   │   ├── FileList/         # File browser/manager
│   │   │   ├── index.jsx
│   │   │   ├── FileItem.jsx
│   │   │   └── styles.css
│   │   ├── MediaPlayer/      # Video/audio player
│   │   │   ├── index.jsx
│   │   │   └── styles.css
│   │   ├── ProgressBar/      # Upload progress indicator
│   │   │   ├── index.jsx
│   │   │   └── styles.css
│   │   └── Auth/             # Login/Register
│   │       ├── Login.jsx
│   │       └── Register.jsx
│   ├── hooks/
│   │   ├── useUpload.js      # Upload logic with progress
│   │   ├── useAuth.js        # Authentication state
│   │   └── useFiles.js       # File listing/management
│   ├── services/
│   │   ├── api.js            # REST API client (axios)
│   │   ├── grpc-web.js       # gRPC-Web client
│   │   └── auth.js           # JWT storage/refresh
│   ├── pages/
│   │   ├── Home.jsx
│   │   ├── Upload.jsx
│   │   ├── Files.jsx
│   │   └── Player.jsx
│   ├── utils/
│   │   ├── formatBytes.js
│   │   └── formatDate.js
│   └── app.jsx               # Main app component
├── public/
│   └── index.html
├── package.json
└── vite.config.js
```
│   │   ├── keygen.go       # Key generation utilities
│   │   └── stream.go       # Streaming encryption
│   ├── api/                 # HTTP REST handlers
│   │   ├── auth.go         # Authentication endpoints
│   │   ├── upload.go       # File upload handler
│   │   ├── download.go     # File download handler
│   │   ├── stream.go       # Media streaming (Range support)
│   │   └── files.go        # File management
│   ├── grpc/                # gRPC services
│   │   ├── server.go       # gRPC server setup
│   │   ├── file_service.go # File metadata operations
│   │   └── admin_service.go # Admin operations
│   ├── storage/             # Storage abstractions
│   │   ├── minio.go        # MinIO client wrapper
│   │   └── redis.go        # Redis client wrapper
│   ├── auth/                # Authentication
│   │   ├── jwt.go          # JWT token generation/validation
│   │   └── middleware.go   # HTTP/gRPC auth middleware
│   ├── worker/              # Background jobs
│   │   ├── cleanup.go      # Auto-delete expired files
│   │   └── scheduler.go    # Job scheduler
│   └── models/              # Data models
│       ├── file.go
│       └── user.go
├── pkg/
│   └── proto/               # Protocol buffers
│       ├── file_service.proto
│       └── admin_service.proto
├── configs/
│   ├── config.yaml          # Server configuration
│   └── docker/              # Docker-specific configs
└── scripts/
    ├── setup-dev.sh         # Development setup
    └── migrate.sh           # Database migrations
```treaming.
- **`internal/grpc`:** Handles metadata, searching, and admin tasks.
- **`internal/worker`:** Background tasks for Auto-Delete cleanup.

### Frontend (Preact)
- **File Manager:** Lists available files.
- **Upload Zone:** Handles Drag-and-Drop and Progress reporting.
- **Media Player:** Standard HTML5 Video/Audio players pointing to stream endpoints.

## Security Considerations
- **TLS is Critical:** Since plaintext is sent to the server, HTTPS/TLS must be enabled at all times to prevent network eavesdropping.
- **Input Validation:** All file names and MIME types are validated on the server to prevent injection attacks.
- **Rate Limiting:** Protects against abuse of the upload/download API.