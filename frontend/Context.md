# File Locker Project - Complete Technical Summary

## Project Overview
**Type**: Encrypted file storage service with Go backend + Preact frontend
**Architecture**: REST API with JWT authentication, AES-256-GCM encryption, MinIO object storage, Redis caching
**Purpose**: Secure file upload/download/streaming with client-side encryption keys

## Technology Stack

### Backend (Go)
- **Framework**: Chi router (go-chi/chi/v5)
- **Storage**: MinIO (port 9012) for encrypted files, Redis (port 6379) for metadata/sessions
- **Encryption**: AES-256-GCM for files at rest, AES-CTR for streaming with HTTP Range support
- **Authentication**: JWT tokens (golang-jwt/jwt), stored in Redis sessions
- **Ports**: HTTP :9010, gRPC :9011
- **Config**: Viper-based config from config.yaml

### Frontend (Preact)
- **Framework**: Preact 10.28.1
- **Build Tool**: Vite 5.0.10
- **Router**: preact-router 4.1.2
- **HTTP Client**: axios 1.13.2 with interceptors for JWT
- **Styling**: Custom CSS with CSS variables, dark/light theme support
- **Port**: Dev server on 5173, proxies /api/v1 to backend :9010

## Complete File Structure

```
lock_files/
├── backend/
│   ├── cmd/server/main.go - Server initialization, routes, handlers
│   ├── internal/
│   │   ├── api/
│   │   │   ├── auth.go - Login/register/logout handlers
│   │   │   ├── upload.go - File upload with multipart form, encryption, MinIO storage
│   │   │   ├── download.go - Full file download with decryption
│   │   │   ├── stream.go - Streaming with Range request support, AES-CTR decryption
│   │   │   ├── files.go - List/search/delete operations
│   │   │   └── utils.go - JSON response helpers
│   │   ├── auth/
│   │   │   ├── jwt.go - Token generation/validation
│   │   │   └── middleware.go - RequireAuth middleware (supports header + query param token)
│   │   ├── config/config.go - Viper configuration loader
│   │   ├── crypto/
│   │   │   └── aes.go - AES-256-GCM encryption/decryption, stream support
│   │   ├── storage/
│   │   │   ├── minio.go - MinIO client, file operations
│   │   │   └── redis.go - Session/metadata/rate-limit operations
│   │   └── worker/cleanup.go - Expired file cleanup worker
│   ├── docs/openapi.yaml - Complete REST API documentation
│   └── go.mod - Dependencies
├── frontend/
│   ├── src/
│   │   ├── main.jsx - App entry point
│   │   ├── app.jsx - Router, global auth state, theme init
│   │   ├── style.css - 856 lines, dark theme, responsive grid, modal styles
│   │   ├── components/
│   │   │   ├── Header.jsx - Nav with theme toggle, user dropdown
│   │   │   ├── FileList.jsx - File display with download/stream/delete, video modal player
│   │   │   ├── FileUpload.jsx - Drag-and-drop upload with progress
│   │   │   └── FileStats.jsx - Stats cards (total files, storage, downloads)
│   │   ├── pages/
│   │   │   ├── Login.jsx - Login form with auth redirect
│   │   │   ├── Register.jsx - Registration form with auth redirect
│   │   │   ├── Dashboard.jsx - 3-column layout (stats|upload|files), session validation
│   │   │   └── Settings.jsx - Placeholder for password change
│   │   └── utils/
│   │       ├── api.js - Axios client with all endpoints, token interceptor
│   │       ├── auth.js - Token/user localStorage management
│   │       └── theme.js - Dark/light theme toggle
│   ├── package.json
│   └── vite.config.js - Proxy config
└── configs/config.yaml
```

## API Endpoints (All at /api/v1)

### Authentication
- `POST /auth/register` - {username, password, email} → JWT token
- `POST /auth/login` - {username, password} → JWT token + session
- `POST /auth/logout` - Invalidates session in Redis

### File Operations
- `POST /upload` - multipart/form-data {file, tags, expire_after} → FileMetadata
- `GET /files` - List all user files → {files: [FileMetadata]}
- `GET /files/search?q=query` - Search by filename/tags
- `DELETE /files?id=fileId` - Delete from MinIO + Redis
- `GET /download/{fileId}` - Download decrypted file
- `GET /stream/{fileId}` - Stream with Range support for video

### Authentication Methods
- **Authorization header**: `Bearer <token>` (used by axios for API calls)
- **Query parameter**: `?token=<token>` (used for download/stream URLs)

## Database Schema (Redis)

### Session Storage
- **Key**: `session:{token}`
- **Value**: `{userID}`
- **TTL**: 1 hour (configurable)

### File Metadata
- **Key**: `file:{fileID}`
- **Value**: JSON `{file_id, user_id, file_name, mime_type, size, created_at, expires_at, tags[], download_count, minio_path, encryption_key}`

### User File Index
- **Key**: `user:files:{userID}`
- **Type**: Set
- **Members**: fileID strings

### Rate Limiting
- **Key**: `ratelimit:{userID}:{window}`
- **Value**: Request count
- **TTL**: Window duration

## Authentication Flow

1. **Registration**: `POST /auth/register` → User created, JWT generated, session stored in Redis
2. **Login**: `POST /auth/login` → Credentials validated, JWT generated, session stored
3. **Token Storage**: Frontend stores JWT in `localStorage.token` and username in `localStorage.user`
4. **API Calls**: Axios interceptor adds `Authorization: Bearer <token>` header
5. **Middleware**: `RequireAuth` validates token, checks Redis session, sets `userID` in context
6. **Session Check**: Dashboard validates token on mount/reload, redirects to login if invalid
7. **401 Handling**: Catch block detects expired tokens, shows message, clears session, redirects

## File Upload Flow

1. User selects file via drag-and-drop or file picker
2. Frontend sends multipart form with file, optional tags, expire_after
3. Backend generates unique file ID (UUID)
4. **Encryption**: Generates 32-byte AES key, encrypts file with AES-256-GCM
5. **Storage**: Uploads encrypted file to MinIO at `{userID}/{fileID}`
6. **Metadata**: Stores FileMetadata in Redis with encryption key
7. **Index**: Adds fileID to user's file set in Redis
8. Response returns file metadata (snake_case fields)

## File Download Flow

1. User clicks download button
2. Frontend creates anchor with `href=getDownloadUrl(fileId)` which includes `?token=jwt`
3. Backend middleware validates token from query parameter
4. Retrieves metadata from Redis, validates ownership
5. Fetches encrypted file from MinIO
6. Decrypts using stored encryption key
7. Streams decrypted bytes with `Content-Disposition: attachment`

## File Streaming Flow (Videos)

1. User clicks play button on video file
2. Frontend opens modal with `<video>` element
3. Video src set to `getStreamUrl(fileId)?token=jwt`
4. Browser makes request (may include Range header for seeking)
5. Backend handles full stream or range request:
   - **Full**: Decrypt entire file, stream to response
   - **Range**: Calculate AES counter offset, decrypt from specific block, send partial content
6. Video player supports seeking/scrubbing via Range requests

## Frontend Layout

### Dashboard (3-Column Grid)
- **Grid**: `1fr 2.5fr 1fr` (stats | upload | files)
- **Responsive**: Collapses to 1 column on mobile (768px)
- **Container**: Full width, no max-width constraint
- **Theme**: Dark/light toggle in header, persists to localStorage

### File List Features
- File icon based on type (video vs document)
- Metadata display: size, upload date, expiration, tags
- Actions: Play (videos only), Download, Delete
- Empty state with SVG illustration
- Modal video player with close button

## Issues Fixed During Development

### Backend Issues
1. **Test Compilation Errors**: Missing mock methods in test files - Fixed by adding complete mock implementations
2. **File Deletion Bug**: Only cleared Redis, not MinIO - Fixed by calling both `minioStorage.DeleteFile()` and Redis cleanup
3. **Token Validation**: Only checked Authorization header - Fixed by adding query parameter support in middleware

### Frontend Issues  
4. **Authentication State**: Username showing when not logged in - Fixed conditional rendering in Header
5. **Login Redirect Loop**: Login/register accessible when authenticated - Added `useEffect` checks with redirect
6. **Files Not Displaying**: Backend returns snake_case, frontend expected camelCase - Updated all components to use snake_case field names (file_id, file_name, created_at)
7. **File Name Display**: Names appearing vertically - Fixed CSS with `display: flex; flex-direction: row` and proper flex properties
8. **Dashboard Width**: Constrained to 1000px - Changed to `width: 100%`
9. **Grid Proportions**: Not matching 1:2.5:1 - Updated to `grid-template-columns: 1fr 2.5fr 1fr`
10. **Session Validation**: No check on page reload - Added token validation in Dashboard `useEffect` and 401 error handling
11. **Download Authorization Error**: Direct links missing headers - Updated middleware to support token in query parameter
12. **Stream Downloads Instead of Playing**: Opening in new window - Created modal with embedded `<video>` player

## Current State

### Working Features ✅
- User registration and login
- JWT authentication with session management
- File upload with encryption (AES-256-GCM)
- File listing and search
- File download with decryption
- Video streaming with seeking/Range support
- File deletion (MinIO + Redis)
- Theme toggle (dark/light)
- Session validation on reload
- 401 error handling with auto-logout
- Responsive 3-column dashboard layout
- Modal video player for streaming

### Configuration
- Backend runs on :9010 (HTTP) and :9011 (gRPC)
- MinIO on :9012, Redis on :6379
- JWT expiration: 1 hour
- Session TTL: 1 hour
- Cleanup worker runs every 60 hours

### Code Quality
- All backend tests compile and pass
- Complete OpenAPI documentation at /swagger/index.html
- Frontend builds successfully (77.43 KB JS + 10.44 KB CSS gzipped)
- No compilation errors or warnings

## Pending Feature Requests (From Latest Message)

### 1. File Description and Rename
- **Requirement**: Add description field to files, store in metadata
- **Rename**: Allow changing file display name (not MinIO path)
- **Storage**: Update FileMetadata structure in Redis
- **API**: New endpoints for update operations
- **CLI Compatibility**: Must work via REST API (gRPC already has proto definitions)

### 2. Export All Files
- **Requirement**: Download all user files as single password-protected tar/zip
- **Implementation**: 
  - Backend endpoint to fetch all files, decrypt each
  - Create tar archive in memory or temp location
  - Password-protect with separate password (user input)
  - Stream archive to client
- **Challenges**: Memory usage for large datasets, streaming tar generation
- **CLI**: Same endpoint accessible via curl/CLI tool

### 3. Password Change
- **Requirement**: Change user login password
- **Implementation**:
  - New endpoint: `POST /auth/change-password` with {old_password, new_password}
  - Validate old password, hash new password
  - Update user record (currently no persistent user DB, uses JWT claims)
  - **Issue**: No user database yet, auth only validates against mock/config
  - **Needs**: User persistence layer (PostgreSQL/MySQL or Redis hash)
- **Frontend**: Settings page already exists, needs form implementation
- **CLI**: Via API call

## Critical Technical Notes

1. **Encryption Keys**: Stored in Redis metadata (not ideal for production - consider key management service)
2. **User Storage**: Currently no persistent user database, only JWT claims and Redis sessions
3. **File Storage**: MinIO paths use `{userID}/{fileID}` structure
4. **Token Expiration**: Frontend handles 401 errors, shows "Session expired" message, auto-redirects after 2s
5. **CORS**: Enabled for frontend development (localhost:5173)
6. **Middleware Order**: CORS → RequireAuth → Handler
7. **Video Streaming**: Uses AES-CTR mode for efficient Range request support
8. **Frontend-Backend Contract**: Backend returns snake_case JSON, frontend uses exact field names

## Next Steps for New Features

1. **Add User Database**: PostgreSQL with bcrypt password hashing
2. **Update FileMetadata**: Add `description` field, `display_name` field
3. **Implement Update Endpoint**: `PATCH /files?id={fileId}` with {description?, display_name?}
4. **Create Export Endpoint**: `POST /export` → password-protected tar archive
5. **Password Change**: `POST /auth/change-password` with DB updates
6. **Frontend Forms**: File edit modal, export dialog, password change in settings
7. **CLI Tool**: Go binary using same API endpoints with flag-based interface

This summary represents the complete current state of the File Locker project as of 2025-12-28.