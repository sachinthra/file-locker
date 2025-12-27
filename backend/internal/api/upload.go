package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sachinthra/file-locker/backend/internal/crypto"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type UploadHandler struct {
	minioStorage *storage.MinIOStorage
	redisCache   *storage.RedisCache
}

func NewUploadHandler(minioStorage *storage.MinIOStorage, redisCache *storage.RedisCache) *UploadHandler {
	return &UploadHandler{
		minioStorage: minioStorage,
		redisCache:   redisCache,
	}
}

type UploadResponse struct {
	FileID        string     `json:"file_id"`
	FileName      string     `json:"file_name"`
	Size          int64      `json:"size"`
	MimeType      string     `json:"mime_type"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	DownloadCount int        `json:"download_count"`
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// 10 MB is plenty for headers and small fields. Large files will stream from disk.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Check file size limit (500MB)
	maxSize := int64(500 << 20)
	if header.Size > maxSize {
		respondError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("File too large. Max size: %d MB", maxSize/(1<<20)))
		return
	}

	// Get optional parameters
	expireAfterStr := r.FormValue("expire_after") // in hours
	tagsStr := r.FormValue("tags")                // comma-separated

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// Parse expiration
	var expiresAt *time.Time
	if expireAfterStr != "" {
		hours, err := strconv.Atoi(expireAfterStr)
		if err == nil && hours > 0 {
			expiry := time.Now().Add(time.Duration(hours) * time.Hour)
			expiresAt = &expiry
		}
	}

	// Generate unique fileID
	fileID := uuid.New().String()

	// Generate encryption key
	key, err := crypto.GenerateKey()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate encryption key")
		return
	}

	// Create encrypted stream
	encryptedReader, err := crypto.EncryptStream(file, key)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to encrypt file")
		return
	}

	// Determine content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// MinIO path
	minioPath := fmt.Sprintf("%s/%s", userID, fileID)

	// Upload to MinIO (encrypted size is original size + IV size)
	encryptedSize := header.Size + 16 // 16 bytes for IV
	err = h.minioStorage.SaveFile(r.Context(), minioPath, encryptedReader, encryptedSize, "application/octet-stream")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	// Encode encryption key for storage
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Create metadata
	metadata := &storage.FileMetadata{
		FileID:        fileID,
		UserID:        userID,
		FileName:      header.Filename,
		MimeType:      contentType,
		Size:          header.Size,
		EncryptedSize: encryptedSize,
		MinIOPath:     minioPath,
		EncryptionKey: encodedKey,
		CreatedAt:     time.Now(),
		ExpiresAt:     expiresAt,
		Tags:          tags,
		DownloadCount: 0,
	}

	// Determine Redis expiration (if file has expiration, use that + buffer)
	var redisExpiration time.Duration
	if expiresAt != nil {
		redisExpiration = time.Until(*expiresAt) + 24*time.Hour
	} else {
		redisExpiration = 0 // No expiration
	}

	// Save metadata to Redis
	if err := h.redisCache.SaveFileMetadata(r.Context(), fileID, metadata, redisExpiration); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save file metadata")
		return
	}

	// Add file to user's index
	if err := h.redisCache.AddFileToUserIndex(r.Context(), userID, fileID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to index file")
		return
	}

	// Return response
	respondJSON(w, http.StatusCreated, UploadResponse{
		FileID:        fileID,
		FileName:      header.Filename,
		Size:          header.Size,
		MimeType:      contentType,
		CreatedAt:     metadata.CreatedAt,
		ExpiresAt:     expiresAt,
		DownloadCount: 0,
	})
}
