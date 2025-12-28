package worker

import (
	"context"
	"log"
	"time"

	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type CleanupWorker struct {
	minioStorage *storage.MinIOStorage
	pgStore      *storage.PostgresStore
	interval     time.Duration
}

func NewCleanupWorker(minio *storage.MinIOStorage, pgStore *storage.PostgresStore, interval time.Duration) *CleanupWorker {
	return &CleanupWorker{
		minioStorage: minio,
		pgStore:      pgStore,
		interval:     interval,
	}
}

func (w *CleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run cleanup immediately on start
	w.cleanup(ctx)

	for {
		select {
		case <-ticker.C:
			w.cleanup(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *CleanupWorker) cleanup(ctx context.Context) {
	// Get expired files from PostgreSQL
	expiredFiles, err := w.pgStore.GetExpiredFiles(ctx)
	if err != nil {
		log.Printf("Failed to get expired files: %v", err)
		return
	}

	if len(expiredFiles) == 0 {
		log.Println("No expired files to clean up")
		return
	}

	filesDeleted := 0
	spaceFreed := int64(0)

	for _, metadata := range expiredFiles {
		// Delete file from MinIO
		if err := w.minioStorage.DeleteFile(ctx, metadata.MinIOPath); err != nil {
			log.Printf("Failed to delete file from MinIO: %s, error: %v", metadata.FileID, err)
			continue
		}

		// Delete metadata from PostgreSQL
		if err := w.pgStore.DeleteFileMetadata(ctx, metadata.FileID); err != nil {
			log.Printf("Failed to delete file metadata: %s, error: %v", metadata.FileID, err)
			continue
		}

		filesDeleted++
		spaceFreed += metadata.Size
	}

	log.Printf("Cleanup completed: %d files deleted, %d bytes freed", filesDeleted, spaceFreed)
}
