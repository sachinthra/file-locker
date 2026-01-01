# File Locker - Developer Guide

## 1. Development Environment (Local)
To run the full stack (Backend + Frontend + DBs) locally with hot-reloading:

```bash
make dev
```

**Access Points:**
* **Backend API:** http://localhost:9010
* **Frontend:** 
  - **Dev Mode (Vite):** http://localhost:5173
  - **Production (Nginx):** http://localhost:80
* **API Documentation:** 
  - **Swagger UI:** 
    - Dev: http://localhost:5173/swagger/index.html
    - Prod: http://localhost/swagger/index.html
  - **Custom Docs:** 
    - Dev: http://localhost:5173/docs
    - Prod: http://localhost/docs
  - **Raw OpenAPI:** 
    - Dev: http://localhost:5173/api/v1/docs/openapi.yaml
    - Prod: http://localhost/api/v1/docs/openapi.yaml

## 2. API Documentation

The File Locker API is documented using OpenAPI 3.0 specification.

### Accessing API Docs (3 Ways):

#### Option 1: Swagger UI (Recommended for Testing)
**Development Mode:**
```
http://localhost:5173/swagger/index.html
```

**Production Mode:**
```
http://localhost/swagger/index.html
```

- Full interactive documentation
- "Try it out" buttons to test endpoints
- Live request/response examples
- Authorization support

#### Option 2: In-App Viewer
**Development Mode:**
```
http://localhost:5173/docs
```

**Production Mode:**
```
http://localhost/docs
```

- Integrated into File Locker frontend
- Accessible from user menu → "API Docs"
- Same Swagger UI, styled for the app

#### Option 3: Raw OpenAPI YAML
**Development Mode:**
```
http://localhost:5173/api/v1/docs/openapi.yaml
```

**Production Mode:**
```
http://localhost/api/v1/docs/openapi.yaml
```

- Download the OpenAPI specification
- Import into Postman/Insomnia/Thunder Client
- Use for code generation tools

### Updating API Documentation

The OpenAPI specification is located at:
```
backend/docs/openapi.yaml
```

After making API changes:
1. Update `backend/docs/openapi.yaml`
2. Restart backend server
3. Refresh Swagger UI to see changes

## 3. Release & Distribution
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
