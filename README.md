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

### Method 2: CLI Usage

```bash
# Upload a file
filelocker upload my-video.mp4

# List files
filelocker list

# Download a file
filelocker download <file-id>

```

## Architecture

File Locker uses a standard **Server-Side Encryption** architecture:

* **Frontend:** Preact app communicating via REST (Files) and gRPC (Metadata).
* **Backend:** Go server handling encryption streams.
* **Storage:** MinIO stores the encrypted blobs.

See [ARCHITECTURE.md](ARCHITECTURE.md) for technical details and [IMPLEMENTATION.md](IMPLEMENTATION.md) for developer guides.

## Roadmap

See [Design.md](https://www.google.com/search?q=Design.md) for the detailed implementation roadmap.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](https://www.google.com/search?q=CONTRIBUTING.md).

## License

This project is licensed under the MIT License - see the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.

---

<div align="center">
Made with 🔐 by [Your Name]
</div>