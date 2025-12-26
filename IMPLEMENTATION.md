# Implementation Details & Developer Guide

This document contains specific instruction sets for developers regarding UI/UX features, complex logic, and specific component implementations.

## 1. Drag-and-Drop Implementation
**Target Component:** `frontend/src/components/FileUpload.jsx`

- **Logic:**
  - Use the HTML5 Drag and Drop API.
  - Listen for `dragover` (prevent default behavior) and `drop`.
  - Access files via `event.dataTransfer.files`.
- **UX Requirements:**
  - Change the border color/style of the drop zone when a user drags a file over it.
  - Immediately show the file list with "Pending" status upon drop.

## 2. Batch Operations (Multiple Files)
**Target Component:** `frontend/src/hooks/useUpload.js`

- **Frontend Logic:**
  - When multiple files are dropped, do **not** combine them into one request.
  - Iterate through the file list and trigger a separate `uploadFile()` function for each.
  - Use `Promise.allSettled()` or a queue system to manage concurrency (e.g., max 3 uploads at a time).
- **Backend Logic:**
  - The API endpoint `POST /api/v1/upload` accepts a single file.
  - This simplifies error handling—if one file fails, the others still succeed.

## 3. Progress Indicators
**Target Component:** `frontend/src/components/ProgressBar.jsx`

- **Implementation:**
  - Use `axios` or `XMLHttpRequest` which provides `onUploadProgress`.
  - Calculate percentage: `Math.round((progressEvent.loaded * 100) / progressEvent.total)`.
- **UI:**
  - Display a progress bar next to *each* file in the batch list.
  - Color codes: Blue (Uploading), Green (Complete), Red (Error).

## 4. Large Video Streaming
**Target Component:** `backend/internal/api/stream.go`

- **Why:** We cannot download a 2GB video to browser memory to decrypt it.
- **How:**
  - The Go server must implement `ServeContent` or handle `Range` headers manually.
  - When MinIO returns the encrypted stream, the Go server reads chunk -> decrypts chunk -> writes chunk to HTTP response.
- **Frontend:**
  - Simply use: `<video controls src="/api/v1/stream/{file_id}" />`.
  - The browser handles buffering and memory management automatically.

## 5. Auto-Delete Logic
**Target Component:** `backend/internal/worker/cleanup.go`

- **Mechanism:**
  - Use a Go `time.Ticker` (e.g., runs every hour).
  - SQL Query: `SELECT id, minio_path FROM files WHERE expires_at < NOW()`.
  - Action: Delete object from MinIO, then delete row from DB.
- **User Option:** Allow users to set "Delete after 1 hour" or "Delete after 1 download" during upload.
