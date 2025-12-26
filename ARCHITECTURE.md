# File Locker - Architecture Documentation

## Table of Contents
- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Security Architecture](#security-architecture)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [Deployment Architecture](#deployment-architecture)
- [Technology Decisions](#technology-decisions)

## Overview

File Locker implements a **zero-knowledge architecture** where all encryption and decryption operations happen on the client side. The server acts as a secure storage and coordination layer but never has access to plaintext data or encryption keys.

### Design Principles

1. **Zero-Knowledge:** Server never accesses plaintext keys or data
2. **Defense in Depth:** Multiple security layers (encryption, authentication, authorization)
3. **Client-Side Security:** All cryptographic operations happen client-side
4. **Secure by Default:** Conservative defaults with opt-in relaxation
5. **Modularity:** Clean separation between components

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                          │
├─────────────────┬─────────────────┬────────────────────────┤
│   Web GUI       │      TUI        │         CLI            │
│   (Preact)      │   (Bubble Tea)  │    (Cobra/Viper)       │
└────────┬────────┴────────┬────────┴───────────┬────────────┘
         │                 │                    │
         └─────────────────┼────────────────────┘
                           │
                    ┌──────▼──────┐
                    │   Crypto    │  ← Client-side encryption
                    │   Module    │    (AES-256-GCM, Argon2id)
                    └──────┬──────┘
                           │
         ┌─────────────────┼────────────────────┐
         │            gRPC + TLS                 │
         │         (Port 9011)                   │
         └─────────────────┬────────────────────┘
                           │
┌──────────────────────────▼────────────────────────────────┐
│                    Server Layer (Go)                       │
├─────────────────┬──────────────────┬─────────────────────┤
│  gRPC Server    │   HTTP Server    │   Auth Middleware   │
│                 │   (Port 9010)    │   (JWT + Rate Limit)│
└────────┬────────┴────────┬─────────┴──────────┬──────────┘
         │                 │                     │
    ┌────▼────┐      ┌────▼────┐          ┌────▼────┐
    │  File   │      │ Storage │          │  User   │
    │ Service │      │ Service │          │ Service │
    └────┬────┘      └────┬────┘          └────┬────┘
         │                │                     │
┌────────┴────────────────┴─────────────────────┴──────────┐
│                   Storage Layer                           │
├───────────────────┬───────────────────┬──────────────────┤
│     MinIO         │      Redis        │   Local FS       │
│  (Port 9012)      │   (Port 9013)     │  (Metadata)      │
│  Encrypted Files  │  Sessions/Cache   │  Config/Logs     │
└───────────────────┴───────────────────┴──────────────────┘
```

## Security Architecture

### Zero-Knowledge Model

```
User's Machine                          Server
┌─────────────┐                    ┌──────────────┐
│             │                    │              │
│  Password   │                    │   Receives   │
│     +       │                    │   Encrypted  │
│  Argon2id   │  ──Encrypted──>    │    Blob      │
│     ↓       │     Data Only      │              │
│  Master Key │                    │   Stores     │
│     ↓       │                    │   in MinIO   │
│  AES-GCM    │                    │              │
│  Encrypt    │                    │  (Cannot     │
│     ↓       │                    │   decrypt)   │
│ Ciphertext  │                    │              │
└─────────────┘                    └──────────────┘
```

### Encryption Flow

1. **Password → Master Key**
   ```
   Master Key = Argon2id(password, salt, params)
   Params: time=3, memory=64MB, parallelism=4
   ```

2. **File Encryption**
   ```
   Ciphertext = AES-256-GCM.Encrypt(plaintext, master_key, nonce)
   HMAC = HMAC-SHA256(ciphertext, master_key)
   Output = ciphertext || HMAC || metadata
   ```

3. **Decryption**
   ```
   Verify HMAC first (tamper detection)
   Plaintext = AES-256-GCM.Decrypt(ciphertext, master_key, nonce)
   ```

### Key Derivation Hierarchy

```
User Password
     │
     ├──> Argon2id ──> Master Key (256-bit)
     │                      │
     │                      ├──> Encryption Key (AES-256)
     │                      ├──> HMAC Key
     │                      └──> Recovery Key Encryption Key
     │
     └──> BIP39 Mnemonic ──> Recovery Key
                                  │
                                  └──> Can regenerate Master Key
```

## Component Details

### Backend (Go)

```
backend/
├── cmd/
│   ├── server/          # Main server entry point
│   └── cli/             # CLI application
├── internal/
│   ├── crypto/          # Encryption/decryption logic
│   │   ├── aes.go      # AES-256-GCM implementation
│   │   ├── argon2.go   # Key derivation
│   │   └── hmac.go     # Integrity verification
│   ├── grpc/           # gRPC server and handlers
│   │   ├── server.go
│   │   ├── file_service.go
│   │   └── user_service.go
│   ├── storage/        # Storage abstractions
│   │   ├── minio.go
│   │   └── local.go
│   ├── auth/           # JWT and authentication
│   │   ├── jwt.go
│   │   └── middleware.go
│   └── models/         # Data models
├── pkg/                # Public packages
│   └── proto/          # Protocol buffers
└── configs/            # Configuration files
```

### Frontend (Preact)

```
frontend/
├── src/
│   ├── components/
│   │   ├── FileUpload/      # Drag-and-drop upload
│   │   ├── FileList/        # Locked files display
│   │   ├── EncryptModal/    # Encryption dialog
│   │   └── DecryptModal/    # Decryption dialog
│   ├── crypto/
│   │   ├── worker.js        # Web Worker for encryption
│   │   ├── aes.js          # AES implementation (WebCrypto)
│   │   └── argon2.wasm     # Argon2 WASM module
│   ├── services/
│   │   ├── grpc-web.js     # gRPC-Web client
│   │   └── auth.js         # JWT handling
│   ├── hooks/              # React hooks
│   └── pages/              # Page components
└── public/
```

### TUI (Terminal User Interface)

```
tui/
├── main.go
├── ui/
│   ├── file_list.go    # File browser view
│   ├── encrypt.go      # Encryption dialog
│   ├── decrypt.go      # Decryption dialog
│   └── settings.go     # Settings view
└── styles/
    └── theme.go        # Color scheme
```

## Data Flow

### Encryption Flow

```
1. User selects file in UI
2. User enters password
   │
3. CLIENT: Derive key using Argon2id
   │
4. CLIENT: Read file into memory
   │
5. CLIENT: Encrypt with AES-256-GCM
   │
6. CLIENT: Generate HMAC
   │
7. CLIENT: Create metadata JSON
   │
8. ─── gRPC/TLS ───> SERVER: Upload encrypted blob
   │
9. SERVER: Store in MinIO
   │
10. SERVER: Update metadata in Redis
   │
11. <─── gRPC/TLS ─── SERVER: Return file ID
   │
12. CLIENT: Display success + file ID
```

### Decryption Flow

```
1. User selects locked file
2. User enters password
   │
3. ─── gRPC/TLS ───> SERVER: Request file
   │
4. SERVER: Retrieve from MinIO
   │
5. <─── gRPC/TLS ─── SERVER: Return encrypted blob
   │
6. CLIENT: Verify HMAC
   │
7. CLIENT: Derive key using Argon2id
   │
8. CLIENT: Decrypt with AES-256-GCM
   │
9. CLIENT: Save decrypted file
   │
10. CLIENT: Schedule auto-delete after inactivity
```

## Deployment Architecture

### Local Development

```
docker-compose.yml
├── filelocker-server    (Port 9010, 9011)
├── minio                (Port 9012, 9014)
└── redis                (Port 9013)
```

### Home Network Deployment

```
┌──────────────────────────────────────────────┐
│            Home Network (192.168.1.0/24)      │
│                                               │
│  ┌────────────┐         ┌─────────────────┐ │
│  │   Router   │ ────────│  File Locker    │ │
│  │ (Gateway)  │         │  192.168.1.100  │ │
│  └──────┬─────┘         │  Ports:         │ │
│         │               │  - 9010 (HTTP)  │ │
│         │               │  - 9011 (gRPC)  │ │
│         │               └─────────────────┘ │
│         │                                    │
│    ┌────┴────┐                              │
│    │ Devices │                              │
│    ├─────────┤                              │
│    │ Laptop  │  http://192.168.1.100:9010   │
│    │ Phone   │                              │
│    │ Tablet  │                              │
│    └─────────┘                              │
└──────────────────────────────────────────────┘
```

### Production Deployment (Future)

```
┌─────────────────────────────────────────────────┐
│                  Load Balancer                   │
│                   (HTTPS/TLS)                    │
└─────────────┬───────────────────┬───────────────┘
              │                   │
    ┌─────────▼─────────┐  ┌─────▼──────────────┐
    │  Server Instance  │  │  Server Instance   │
    │     (Replica)     │  │     (Replica)      │
    └─────────┬─────────┘  └─────┬──────────────┘
              │                   │
    ┌─────────▼───────────────────▼──────────────┐
    │           MinIO Cluster (S3)               │
    │         (Replicated Storage)               │
    └────────────────────────────────────────────┘
    ┌────────────────────────────────────────────┐
    │         Redis Cluster (Sentinel)           │
    │         (Session Management)               │
    └────────────────────────────────────────────┘
```

## Technology Decisions

### Why Go for Backend?

- **Performance:** Excellent for I/O operations
- **Concurrency:** Goroutines for handling multiple requests
- **Crypto Library:** Robust `crypto` standard library
- **gRPC Support:** First-class gRPC support
- **Binary Distribution:** Single binary deployment

### Why Preact for Frontend?

- **Small Size:** 3KB gzipped (vs React 40KB)
- **Performance:** Fast rendering
- **React Compatibility:** Use React ecosystem
- **WebCrypto:** Native browser crypto APIs

### Why gRPC?

- **Performance:** Binary protocol, faster than JSON
- **Streaming:** Bidirectional streaming for large files
- **Type Safety:** Protocol buffers provide strong typing
- **Code Generation:** Auto-generate client/server code

### Why MinIO?

- **S3 Compatible:** Industry-standard API
- **Self-Hosted:** Full control over data
- **Scalable:** Can scale from single node to cluster
- **Performant:** Optimized for object storage

### Why Redis?

- **Fast:** In-memory cache for session data
- **Simple:** Easy to set up and use
- **Persistence:** Optional disk persistence
- **Pub/Sub:** For real-time notifications

## Security Considerations

### Threat Model

**In Scope:**
- Password brute force attacks
- Network eavesdropping
- Server compromise
- Data tampering

**Out of Scope (User Responsibility):**
- Client machine compromise
- Physical access to unlocked files
- Social engineering attacks

### Mitigation Strategies

1. **Brute Force Protection**
   - Rate limiting (10 attempts per minute)
   - Account lockout after 5 failed attempts
   - Argon2id makes each attempt expensive

2. **Network Attacks**
   - TLS 1.3 for all communication
   - Certificate pinning (optional)
   - No plaintext transmission

3. **Server Compromise**
   - Zero-knowledge: Server can't decrypt data
   - Encrypted blobs in storage
   - No key material on server

4. **Data Tampering**
   - HMAC verification before decryption
   - Detect any modifications to ciphertext

## Performance Considerations

### File Size Limits

- **Small Files (<10MB):** In-memory processing
- **Large Files (>10MB):** Streaming encryption
- **Very Large Files (>1GB):** Chunked processing with progress

### Optimization Strategies

1. **Chunked Encryption:** Process large files in chunks
2. **Web Workers:** Offload crypto to separate thread
3. **Connection Pooling:** Reuse gRPC connections
4. **Redis Caching:** Cache metadata lookups
5. **Compression:** Optional pre-encryption compression

## Monitoring & Observability

### Metrics (Prometheus)

```
# [To be filled - specific metrics]
- file_encryption_duration_seconds
- file_decryption_duration_seconds
- grpc_request_duration_seconds
- storage_operations_total
- auth_attempts_total
```

### Logging

```
# [To be filled - logging strategy]
- Structured logging (JSON)
- Log levels: DEBUG, INFO, WARN, ERROR
- No sensitive data in logs
```

### Tracing (Future)

- [To be filled - OpenTelemetry integration]

## Future Considerations

### Scalability

- Horizontal scaling of server instances
- MinIO distributed mode
- Redis Sentinel for HA

### Mobile Apps

- Native iOS/Android apps
- Same gRPC backend
- Platform-specific crypto APIs

### Browser Extensions

- Quick encrypt/decrypt from context menu
- Integration with file managers

---

*Last Updated: [To be filled]*
*Version: 0.1.0 (Planning Phase)*
