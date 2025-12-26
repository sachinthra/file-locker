package api

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type FilesHandler struct {
	redisCache *storage.RedisCache
}

func NewFilesHandler(redisCache *storage.RedisCache) *FilesHandler {
	return &FilesHandler{
		redisCache: redisCache,
	}
}

type FileInfo struct {
	FileID        string     `json:"file_id"`
	FileName      string     `json:"file_name"`
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

	// Get list of fileIDs from Redis user index
	fileIDs, err := h.redisCache.GetUserFiles(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve files")
		return
	}

	// Collect file metadata
	files := make([]FileInfo, 0)
	now := time.Now()

	for _, fileID := range fileIDs {
		metadata, err := h.redisCache.GetFileMetadata(r.Context(), fileID)
		if err != nil {
			// Skip files that no longer exist
			continue
		}

		// Filter out expired files
		if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(now) {
			continue
		}

		files = append(files, FileInfo{
			FileID:        metadata.FileID,
			FileName:      metadata.FileName,
			MimeType:      metadata.MimeType,
			Size:          metadata.Size,
			CreatedAt:     metadata.CreatedAt,
			ExpiresAt:     metadata.ExpiresAt,
			Tags:          metadata.Tags,
			DownloadCount: metadata.DownloadCount,
		})
	}

	// Sort by created_at (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].CreatedAt.After(files[j].CreatedAt)
	})

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

	query = strings.ToLower(query)

	// Get user's files
	fileIDs, err := h.redisCache.GetUserFiles(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve files")
		return
	}

	// Filter matching files
	matchingFiles := make([]FileInfo, 0)
	now := time.Now()

	for _, fileID := range fileIDs {
		metadata, err := h.redisCache.GetFileMetadata(r.Context(), fileID)
		if err != nil {
			continue
		}

		// Skip expired files
		if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(now) {
			continue
		}

		// Check if filename matches
		if strings.Contains(strings.ToLower(metadata.FileName), query) {
			matchingFiles = append(matchingFiles, FileInfo{
				FileID:        metadata.FileID,
				FileName:      metadata.FileName,
				MimeType:      metadata.MimeType,
				Size:          metadata.Size,
				CreatedAt:     metadata.CreatedAt,
				ExpiresAt:     metadata.ExpiresAt,
				Tags:          metadata.Tags,
				DownloadCount: metadata.DownloadCount,
			})
			continue
		}

		// Check if any tag matches
		for _, tag := range metadata.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matchingFiles = append(matchingFiles, FileInfo{
					FileID:        metadata.FileID,
					FileName:      metadata.FileName,
					MimeType:      metadata.MimeType,
					Size:          metadata.Size,
					CreatedAt:     metadata.CreatedAt,
					ExpiresAt:     metadata.ExpiresAt,
					Tags:          metadata.Tags,
					DownloadCount: metadata.DownloadCount,
				})
				break
			}
		}
	}

	// Sort by relevance (for now, just by created_at)
	sort.Slice(matchingFiles, func(i, j int) bool {
		return matchingFiles[i].CreatedAt.After(matchingFiles[j].CreatedAt)
	})

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
	metadata, err := h.redisCache.GetFileMetadata(r.Context(), fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	// Verify ownership
	if metadata.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Delete metadata from Redis
	if err := h.redisCache.DeleteFileMetadata(r.Context(), fileID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete file metadata")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "File deleted successfully",
		"file_id": fileID,
	})
}
