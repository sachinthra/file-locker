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

## Data Flow

### Upload (Encryption)
1. **User** drags file to Web UI.
2. **Client** uploads file via HTTP `POST /upload` (using Multipart or Binary stream).
3. **Server** authenticates the request via JWT.
4. **Server** generates a unique Data Key for the file.
5. **Server** streams the upload through an AES-256-GCM encrypter.
6. **Server** saves the *Encrypted* stream to MinIO.
7. **Server** saves metadata (Filename, Key ID, Size) to Redis/DB.

### Download / Streaming (Decryption)
1. **User** requests file `GET /download/{id}` or `<video src="/stream/{id}">`.
2. **Server** authenticates user.
3. **Server** retrieves Encrypted stream from MinIO.
4. **Server** decrypts the stream on-the-fly.
5. **Client** receives plaintext stream.
   - *Note:* For videos, the server supports HTTP `Range` requests to allow seeking.

## Component Details

### Backend (Go)
- **`internal/crypto`:** Handles AES encryption/decryption logic.
- **`internal/api`:** HTTP handlers for file upload/download and streaming.
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