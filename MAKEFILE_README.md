# File Locker - Developer Guide

## 1. Development Environment (Local)
To run the full stack (Backend + Frontend + DBs) locally with hot-reloading:

```bash
make dev
```

* **Backend:** http://localhost:9010
* **Frontend:** http://localhost:80

## 2. Release & Distribution
To build the artifacts for all 4 environments (Ubuntu, Pi, Mac, Windows):

```bash
export DOCKER_USER="yourusername"
make build-release
```

### Output Artifacts
| Environment | Artifact Location | Description |
| :--- | :--- | :--- |
| **Ubuntu / Pi** | `Docker Hub` | Images pushed automatically. |
| **Ubuntu / Pi** | `dist/deb/` | The installer files (Compose + Setup script). |
| **Mac (Client)** | `bin/fl-darwin-arm64` | Native CLI for M1/M2 Macs. |
| **Windows (Client)** | `bin/fl-windows-amd64.exe` | Native CLI for Windows. |
