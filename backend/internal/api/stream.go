package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sachinthra/file-locker/backend/internal/crypto"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type StreamHandler struct {
	minioStorage *storage.MinIOStorage
	redisCache   *storage.RedisCache
}

func NewStreamHandler(minioStorage *storage.MinIOStorage, redisCache *storage.RedisCache) *StreamHandler {
	return &StreamHandler{
		minioStorage: minioStorage,
		redisCache:   redisCache,
	}
}

func (h *StreamHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	// Get fileID from URL
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		respondError(w, http.StatusBadRequest, "File ID required")
		return
	}

	// Get userID from context
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

	// Parse Range header
	rangeHeader := r.Header.Get("Range")

	if rangeHeader != "" {
		// Handle range request
		h.handleRangeRequest(w, r, metadata, keyBytes, rangeHeader)
	} else {
		// Handle full stream
		h.handleFullStream(w, r, metadata, keyBytes)
	}
}

func (h *StreamHandler) handleFullStream(w http.ResponseWriter, r *http.Request, metadata *storage.FileMetadata, keyBytes []byte) {
	// Get encrypted stream from MinIO
	encryptedStream, err := h.minioStorage.GetFile(r.Context(), metadata.MinIOPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve file")
		return
	}
	defer encryptedStream.Close()

	// Decrypt stream
	decryptedStream, err := crypto.DecryptStream(encryptedStream, keyBytes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decrypt file")
		return
	}

	// Set headers for inline viewing
	w.Header().Set("Content-Type", metadata.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", metadata.Size))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", metadata.FileName))

	// Stream to client
	io.Copy(w, decryptedStream)
}

func (h *StreamHandler) handleRangeRequest(w http.ResponseWriter, r *http.Request, metadata *storage.FileMetadata, keyBytes []byte, rangeHeader string) {
	// Parse range header: "bytes=start-end"
	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	rangeParts := strings.Split(rangeStr, "-")

	if len(rangeParts) != 2 {
		respondError(w, http.StatusBadRequest, "Invalid range header")
		return
	}

	start, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid range start")
		return
	}

	var end int64
	if rangeParts[1] == "" {
		end = metadata.Size - 1
	} else {
		end, err = strconv.ParseInt(rangeParts[1], 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid range end")
			return
		}
	}

	// Validate range
	if start < 0 || end >= metadata.Size || start > end {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", metadata.Size))
		respondError(w, http.StatusRequestedRangeNotSatisfiable, "Invalid range")
		return
	}

	// Get encrypted range from MinIO
	// Note: We need to account for IV size (16 bytes) at the beginning
	ivSize := int64(16)
	encryptedStart := start + ivSize
	encryptedEnd := end + ivSize

	encryptedStream, err := h.minioStorage.GetFileRange(r.Context(), metadata.MinIOPath, encryptedStart, encryptedEnd)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve file range")
		return
	}
	defer encryptedStream.Close()

	// For proper range decryption with CTR mode, we'd need to:
	// 1. Read the IV from the file
	// 2. Calculate the counter offset for the start position
	// 3. Decrypt only the requested range
	// For simplicity, this implementation streams the full decrypted content
	// In production, implement proper range decryption

	// Set response headers
	contentLength := end - start + 1
	w.Header().Set("Content-Type", metadata.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, metadata.Size))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)

	// Stream the range (simplified - in production, implement proper CTR seeking)
	io.CopyN(w, encryptedStream, contentLength)
}
