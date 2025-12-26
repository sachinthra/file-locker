# File Locker

A secure file encryption utility that protects sensitive data using industry-standard encryption algorithms. File Locker provides both command-line and graphical interfaces for encrypting and decrypting files and folders.

## Core Features
- Password-based file and folder encryption/decryption
- Recovery key mechanism for password recovery
- Uses AES-256-GCM encryption with key derivation functions (Argon2id)
- Multiple interface options: GUI (Preact) and TUI/CLI
- Cross-platform support via Docker containerization
- Supports all file formats and directory structures

## Technology Stack
- **Backend:** Go (encryption, file handling, gRPC services)
- **Frontend:** Preact (web interface)
- **Communication:** gRPC with TLS and JWT authentication
- **Containerization:** Docker & Docker Compose
- **Storage:** MinIO (Running locally)
- **Caching:** Redis (session management, metadata caching)

## Implementation Roadmap

### Phase 1: Core Security
- [ ] Implement AES-256-GCM encryption with Argon2id key derivation
- [ ] Add HMAC integrity verification to detect tampering
- [ ] Implement secure memory handling (zero sensitive data after use)
- [ ] Client-side encryption before transmission to server (zero-knowledge architecture)
- [ ] Rate limiting for brute force protection
- [ ] BIP39 mnemonic recovery phrase system

### Phase 2: User Interface & Experience
- [ ] Build TUI for command-line interface
- [ ] Develop Preact-based web GUI with drag-and-drop support
- [ ] Progress indicators for large file operations
- [ ] Batch operations for multiple files/folders
- [ ] Search functionality for locked files by name and tags
- [ ] Metadata preview without full decryption
- [ ] Auto-delete decrypted files after inactivity timeout (Auto-delete decrypted files (CLI/TUI only) and clear browser memory/cache (Web GUI) after inactivity.)

### Phase 3: Advanced Features
- [ ] File versioning system for encrypted file history
- [ ] Secure file sharing with time-limited, password-protected ZIP/TAR bundles
- [ ] Audit logs for access history tracking
- [ ] Two-factor authentication for recovery operations
- [ ] Backup and recovery mechanism (download as encrypted bundle)
- [ ] Automatic key rotation mechanism
- [ ] Hardware security key support (YubiKey)

### Phase 4: Infrastructure & DevOps
- [ ] Docker Compose configuration for easy deployment
- [ ] Secure gRPC communication with TLS and JWT authentication
- [ ] MinIO integration for optional cloud storage
- [ ] Redis for session management and metadata caching
- [ ] Comprehensive unit and integration tests
- [ ] CI/CD pipeline with automated testing
- [ ] Monitoring stack (Prometheus + Grafana)
- [ ] Secret management (Vault or AWS Secrets Manager)
- [ ] gRPC API documentation and schema

### Phase 5: Compliance & Advanced Security
- [ ] FIPS 140-2 compliance for encryption standards
- [ ] WebAssembly implementation for browser-based encryption
- [ ] Steganography option for hiding encrypted files

## Security Model
- **Zero-Knowledge:** Server never accesses plaintext keys or data
- **Client-Side Encryption:** All encryption/decryption happens on client
- **Secure Transport:** TLS 1.3 for all network communication
- **Defense in Depth:** Multiple layers of security (encryption, authentication, authorization)
