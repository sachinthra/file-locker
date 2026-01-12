package api

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/sachinthra/file-locker/backend/internal/constants"
	"github.com/sachinthra/file-locker/backend/internal/crypto"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type ExportHandler struct {
	minioStorage *storage.MinIOStorage
	pgStore      *storage.PostgresStore
}

func NewExportHandler(minioStorage *storage.MinIOStorage, pgStore *storage.PostgresStore) *ExportHandler {
	return &ExportHandler{
		minioStorage: minioStorage,
		pgStore:      pgStore,
	}
}

// HandleExportAll exports all user files as a ZIP archive
func (h *ExportHandler) HandleExportAll(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(constants.UserIDKey).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("[INFO] Export all files requested by user: %s", userID)

	// Get all user files from PostgreSQL
	files, err := h.pgStore.ListUserFiles(r.Context(), userID)
	if err != nil {
		log.Printf("[ERROR] Failed to list user files for export: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve files")
		return
	}

	if len(files) == 0 {
		respondError(w, http.StatusNotFound, "No files to export")
		return
	}

	log.Printf("[INFO] Found %d files to export for user: %s", len(files), userID)

	// Set response headers for ZIP download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"filelocker-export-%s.zip\"", userID[:8]))
	w.WriteHeader(http.StatusOK)

	// Create ZIP writer that writes directly to response
	zipWriter := zip.NewWriter(w)
	defer func() { _ = zipWriter.Close() }()

	successCount := 0
	failCount := 0

	// Process each file
	for _, metadata := range files {
		log.Printf("[DEBUG] Exporting file: %s (ID: %s)", metadata.FileName, metadata.FileID)

		// Download encrypted file from MinIO
		encryptedReader, err := h.minioStorage.GetFile(r.Context(), metadata.MinIOPath)
		if err != nil {
			log.Printf("[ERROR] Failed to download file %s from MinIO: %v", metadata.FileID, err)
			failCount++
			continue
		}

		// Decode encryption key
		key, err := base64.StdEncoding.DecodeString(metadata.EncryptionKey)
		if err != nil {
			log.Printf("[ERROR] Failed to decode encryption key for file %s: %v", metadata.FileID, err)
			defer func() { _ = encryptedReader.Close() }()
			failCount++
			continue
		}

		// Decrypt the file stream
		decryptedReader, err := crypto.DecryptStream(encryptedReader, key)
		if err != nil {
			log.Printf("[ERROR] Failed to decrypt file %s: %v", metadata.FileID, err)
			defer func() { _ = encryptedReader.Close() }()
			failCount++
			continue
		}

		// Create a sanitized filename (avoid path traversal)
		safeFileName := filepath.Base(metadata.FileName)

		// Create entry in ZIP
		zipFileWriter, err := zipWriter.Create(safeFileName)
		if err != nil {
			log.Printf("[ERROR] Failed to create ZIP entry for file %s: %v", metadata.FileID, err)
			defer func() { _ = encryptedReader.Close() }()
			failCount++
			continue
		}

		// Copy decrypted data to ZIP
		written, err := io.Copy(zipFileWriter, decryptedReader)
		defer func() { _ = encryptedReader.Close() }()

		if err != nil {
			log.Printf("[ERROR] Failed to write file %s to ZIP: %v", metadata.FileID, err)
			failCount++
			continue
		}

		log.Printf("[DEBUG] Successfully exported file %s (%d bytes)", metadata.FileName, written)
		successCount++
	}

	// Add a README file with export info
	readmeContent := fmt.Sprintf(
		"FileLocker Export\n"+
			"================\n"+
			"Total Files: %d\n"+
			"Successfully Exported: %d\n"+
			"Failed: %d\n"+
			"\nAll files have been decrypted and are ready to use.\n",
		len(files), successCount, failCount,
	)

	readmeWriter, err := zipWriter.Create("README.txt")
	if err == nil {
		_, _ = readmeWriter.Write([]byte(readmeContent))
	}

	log.Printf("[INFO] Export completed for user %s: %d success, %d failed", userID, successCount, failCount)
}
