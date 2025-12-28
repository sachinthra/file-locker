package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type FilesHandler struct {
	redisCache   *storage.RedisCache
	minioStorage *storage.MinIOStorage
	pgStore      *storage.PostgresStore
}

func NewFilesHandler(redisCache *storage.RedisCache, minioStorage *storage.MinIOStorage, pgStore *storage.PostgresStore) *FilesHandler {
	return &FilesHandler{
		redisCache:   redisCache,
		minioStorage: minioStorage,
		pgStore:      pgStore,
	}
}

type FileInfo struct {
	FileID        string     `json:"file_id"`
	FileName      string     `json:"file_name"`
	DisplayName   string     `json:"display_name,omitempty"`
	Description   string     `json:"description,omitempty"`
	MimeType      string     `json:"mime_type"`
	Size          int64      `json:"size"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	DownloadCount int        `json:"download_count"`
}

func (h *FilesHandler) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get files from PostgreSQL
	metadataList, err := h.pgStore.ListUserFiles(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve files")
		return
	}

	// Convert to FileInfo and filter expired files
	files := make([]FileInfo, 0)
	now := time.Now()

	for _, metadata := range metadataList {
		// Filter out expired files
		if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(now) {
			continue
		}

		files = append(files, FileInfo{
			FileID:        metadata.FileID,
			FileName:      metadata.FileName,
			DisplayName:   metadata.DisplayName,
			Description:   metadata.Description,
			MimeType:      metadata.MimeType,
			Size:          metadata.Size,
			CreatedAt:     metadata.CreatedAt,
			ExpiresAt:     metadata.ExpiresAt,
			Tags:          metadata.Tags,
			DownloadCount: metadata.DownloadCount,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

func (h *FilesHandler) HandleSearchFiles(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get search query from URL parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query required")
		return
	}

	// Search files in PostgreSQL
	metadataList, err := h.pgStore.SearchFiles(r.Context(), userID, query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search files")
		return
	}

	// Convert to FileInfo and filter expired files
	matchingFiles := make([]FileInfo, 0)
	now := time.Now()

	for _, metadata := range metadataList {
		// Skip expired files
		if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(now) {
			continue
		}

		matchingFiles = append(matchingFiles, FileInfo{
			FileID:        metadata.FileID,
			FileName:      metadata.FileName,
			DisplayName:   metadata.DisplayName,
			Description:   metadata.Description,
			MimeType:      metadata.MimeType,
			Size:          metadata.Size,
			CreatedAt:     metadata.CreatedAt,
			ExpiresAt:     metadata.ExpiresAt,
			Tags:          metadata.Tags,
			DownloadCount: metadata.DownloadCount,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"files": matchingFiles,
		"count": len(matchingFiles),
		"query": query,
	})
}

func (h *FilesHandler) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get fileID from URL
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		respondError(w, http.StatusBadRequest, "File ID required")
		return
	}

	// Get metadata to verify ownership
	metadata, err := h.pgStore.GetFileMetadata(r.Context(), fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	// Verify ownership
	if metadata.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Delete file from MinIO storage
	if err := h.minioStorage.DeleteFile(r.Context(), metadata.MinIOPath); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete file from storage")
		return
	}

	// Delete metadata from PostgreSQL
	if err := h.pgStore.DeleteFileMetadata(r.Context(), fileID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete file metadata")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "File deleted successfully",
		"file_id": fileID,
	})
}

type UpdateFileRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

func (h *FilesHandler) HandleUpdateFile(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get fileID from URL
	fileID := chi.URLParam(r, "fileID")
	if fileID == "" {
		respondError(w, http.StatusBadRequest, "File ID required")
		return
	}

	// Parse request body
	var req UpdateFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing metadata to verify ownership
	metadata, err := h.pgStore.GetFileMetadata(r.Context(), fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	// Verify ownership
	if metadata.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Update metadata in PostgreSQL
	if err := h.pgStore.UpdateFileMetadata(r.Context(), fileID, req.DisplayName, req.Description); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update file metadata")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "File updated successfully",
		"file_id":      fileID,
		"display_name": req.DisplayName,
		"description":  req.Description,
	})
}
