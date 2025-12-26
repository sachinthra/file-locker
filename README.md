# File Locker

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED?logo=docker)](https://docker.com)

**A secure file encryption utility with zero-knowledge architecture**

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#documentation)

</div>

---

## Overview

File Locker is a secure file encryption utility that protects your sensitive data using industry-standard AES-256-GCM encryption. It features a zero-knowledge architecture where the server never accesses your plaintext keys or data, ensuring maximum privacy.

**⚠️ Project Status:** Early Development / Planning Phase

## Features

### Current Features
- [To be filled as features are implemented]

### Planned Features
- 🔐 Password-based file and folder encryption (AES-256-GCM)
- 🔑 BIP39 mnemonic recovery phrase system
- 🖥️ Multiple interfaces: Web GUI, TUI, and CLI
- 📦 Batch operations for multiple files
- 🔄 File versioning and history
- 🔗 Secure file sharing with time-limited access
- 🛡️ Zero-knowledge architecture
- 🐳 Docker containerization for easy deployment
- 📊 Audit logs and access history

See [Design.md](Design.md) for the complete feature roadmap.

## Installation

### Prerequisites

- **Docker:** 20.10 or higher
- **Docker Compose:** 2.0 or higher
- **For building from source:**
  - Go 1.21 or higher
  - Node.js 18 or higher

### Method 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/[username]/file-locker.git
cd file-locker

# Start all services
docker-compose up -d

# Access the web interface
open http://localhost:9010
```

### Method 2: Debian Package

```bash
# Download the .deb package
wget [To be filled - release URL]

# Install the package (includes Docker setup)
sudo dpkg -i file-locker_[version]_amd64.deb

# Start the service
sudo systemctl start file-locker

# Access the web interface
open http://localhost:9010
```

### Method 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/[username]/file-locker.git
cd file-locker

# Build the backend
cd backend
go build -o filelocker ./cmd/server

# Build the frontend
cd ../frontend
npm install
npm run build

# Run with Docker Compose
cd ..
docker-compose up -d
```

## Quick Start

### Command Line Interface (CLI)

```bash
# Lock a file
filelocker lock myfile.txt --password
# or use the short alias
fl lock myfile.txt --password

# Lock with recovery phrase
filelocker lock myfile.txt --password --generate-recovery

# Unlock a file
filelocker unlock myfile.txt.locked

# Lock an entire folder
filelocker lock myfolder/ --password --recursive

# Batch lock multiple files
filelocker lock file1.txt file2.pdf file3.jpg --password

# View locked files catalog
filelocker list

# Search locked files
filelocker search "document"
```

### Terminal User Interface (TUI)

```bash
# Launch the interactive TUI
filelocker tui
# or
fl tui
```

### Web Interface

1. Start the services: `docker-compose up -d`
2. Open browser: `http://localhost:9010`
3. [To be filled - UI walkthrough with screenshots]

## Configuration

### Environment Variables

```bash
# Server Configuration
FILELOCKER_PORT=9010              # Web UI port
FILELOCKER_GRPC_PORT=9011         # gRPC server port
FILELOCKER_HOST=0.0.0.0           # Bind address (use 127.0.0.1 for local-only)

# Storage Configuration
MINIO_ENDPOINT=localhost:9012     # MinIO endpoint
MINIO_ROOT_USER=admin             # MinIO username
MINIO_ROOT_PASSWORD=changeme      # MinIO password

# Redis Configuration
REDIS_ADDR=localhost:9013         # Redis address
REDIS_PASSWORD=                   # Redis password (optional)

# Security
JWT_SECRET=your-secret-key        # JWT signing key (generate securely!)
SESSION_TIMEOUT=3600              # Session timeout in seconds
```

### Configuration File

Create `~/.filelocker/config.yaml`:

```yaml
# [To be filled - YAML configuration schema]
server:
  port: 9010
  host: "0.0.0.0"

encryption:
  algorithm: "AES-256-GCM"
  kdf: "Argon2id"

storage:
  path: "~/.filelocker/vault"
  
auto_delete:
  enabled: true
  timeout: 300  # seconds of inactivity
```

## Architecture

File Locker uses a client-server architecture with zero-knowledge encryption:

- **Backend:** Go-based gRPC server handling file operations
- **Frontend:** Preact web application for GUI
- **Storage:** MinIO for encrypted file storage
- **Cache:** Redis for session management
- **Security:** All encryption happens client-side; server never sees plaintext

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed technical documentation.

## Network Deployment

File Locker can run in two modes:

### Local-Only Mode
```bash
# Set host to localhost only
FILELOCKER_HOST=127.0.0.1 docker-compose up -d
```

### Home Network Mode
```bash
# Allow network access
FILELOCKER_HOST=0.0.0.0 docker-compose up -d

# Access from other devices on your network
# http://[your-machine-ip]:9010
```

**Security Note:** When exposing to your home network, ensure you:
- Use strong JWT secrets
- Enable TLS/HTTPS in production
- Configure firewall rules appropriately

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Check what's using the port
lsof -i :9010

# Change the port in docker-compose.yml or use environment variable
FILELOCKER_PORT=9015 docker-compose up -d
```

**Docker permission denied:**
```bash
# Add your user to docker group
sudo usermod -aG docker $USER
# Log out and back in
```

**[To be filled - Additional troubleshooting as issues are discovered]**

## Documentation

- [Design.md](Design.md) - Project design and implementation roadmap
- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical architecture details
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contributing guidelines
- [API Documentation] - [To be filled - gRPC API documentation]

## Security

### Reporting Vulnerabilities

If you discover a security vulnerability, please email [To be filled - security contact]. Do not create a public issue.

### Security Features

- **Zero-Knowledge Architecture:** Server never accesses plaintext keys
- **AES-256-GCM Encryption:** Industry-standard encryption
- **Argon2id KDF:** Secure password-based key derivation
- **TLS 1.3:** Encrypted communication
- **HMAC Integrity:** Tamper detection

See [SECURITY.md] for our security policy [To be filled].

## Roadmap

See [Design.md](Design.md) for the detailed implementation roadmap across 5 phases.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# [To be filled - Development environment setup]
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [To be filled - Credits and acknowledgments]

## Contact

- GitHub: [@[username]](https://github.com/[username])
- Issues: [GitHub Issues](https://github.com/[username]/file-locker/issues)
- Email: [To be filled]

---

<div align="center">
Made with 🔐 by [username]
</div>
