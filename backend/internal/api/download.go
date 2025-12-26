package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sachinthra/file-locker/backend/internal/crypto"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type DownloadHandler struct {
	minioStorage *storage.MinIOStorage
	redisCache   *storage.RedisCache
}

func NewDownloadHandler(minioStorage *storage.MinIOStorage, redisCache *storage.RedisCache) *DownloadHandler {
	return &DownloadHandler{
		minioStorage: minioStorage,
		redisCache:   redisCache,
	}
}

func (h *DownloadHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	// Get fileID from URL
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		respondError(w, http.StatusBadRequest, "File ID required")
		return
	}

	// Get userID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get metadata from Redis
	metadata, err := h.redisCache.GetFileMetadata(r.Context(), fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	// Ownership check
	if metadata.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Check if file is expired
	if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(time.Now()) {
		respondError(w, http.StatusGone, "File has expired")
		return
	}

	// Decode encryption key
	keyBytes, err := base64.StdEncoding.DecodeString(metadata.EncryptionKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode encryption key")
		return
	}

	// Get encrypted stream from MinIO
	encryptedStream, err := h.minioStorage.GetFile(r.Context(), metadata.MinIOPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve file from storage")
		return
	}
	defer encryptedStream.Close()

	// Decrypt stream
	decryptedStream, err := crypto.DecryptStream(encryptedStream, keyBytes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decrypt file")
		return
	}

	// Set response headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", metadata.FileName))
	w.Header().Set("Content-Type", metadata.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", metadata.Size))

	// Stream to client
	if _, err := io.Copy(w, decryptedStream); err != nil {
		// Log error but can't send response as headers already sent
		return
	}

	// Increment download counter (fire and forget)
	go func() {
		metadata.DownloadCount++
		h.redisCache.SaveFileMetadata(r.Context(), fileID, metadata, 0)
	}()
}
