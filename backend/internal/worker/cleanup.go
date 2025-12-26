package worker

import (
	"context"
	"time"

	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type CleanupWorker struct {
	minioStorage *storage.MinIOStorage
	redisCache   *storage.RedisCache
	interval     time.Duration
}

func NewCleanupWorker(minio *storage.MinIOStorage, redis *storage.RedisCache, interval time.Duration) *CleanupWorker {
	return &CleanupWorker{
		minioStorage: minio,
		redisCache:   redis,
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
	// Note: This is a simplified implementation
	// In production, you would:
	// 1. Use Redis SCAN to iterate through all "file:*" keys
	// 2. Check expiration for each file
	// 3. Delete expired files from both MinIO and Redis
	// 4. Track metrics and log results

	// For MVP, we rely on Redis TTL to auto-expire file metadata
	// The cleanup worker can be enhanced later to:
	// - Scan all file keys using Redis SCAN command
	// - Delete corresponding MinIO objects for expired files
	// - Maintain cleanup metrics

	// Example implementation would look like:
	// now := time.Now()
	// filesDeleted := 0
	// spaceFreed := int64(0)
	//
	// iter := w.redisCache.client.Scan(ctx, 0, "file:*", 0).Iterator()
	// for iter.Next(ctx) {
	//     fileKey := iter.Val()
	//     fileID := strings.TrimPrefix(fileKey, "file:")
	//     metadata, err := w.redisCache.GetFileMetadata(ctx, fileID)
	//     if err != nil {
	//         continue
	//     }
	//     if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(now) {
	//         w.minioStorage.DeleteFile(ctx, metadata.MinIOPath)
	//         w.redisCache.DeleteFileMetadata(ctx, fileID)
	//         filesDeleted++
	//         spaceFreed += metadata.Size
	//     }
	// }
}
