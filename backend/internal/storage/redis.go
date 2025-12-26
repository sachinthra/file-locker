package storage

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

// FileMetadata represents file metadata stored in Redis
type FileMetadata struct {
	FileID        string     `json:"file_id"`
	UserID        string     `json:"user_id"`
	FileName      string     `json:"file_name"`
	MimeType      string     `json:"mime_type"`
	Size          int64      `json:"size"`
	EncryptedSize int64      `json:"encrypted_size"`
	MinIOPath     string     `json:"minio_path"`
	EncryptionKey string     `json:"encryption_key"` // Base64 encoded (In production, use separate secure storage)
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	DownloadCount int        `json:"download_count"`
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	// 1. Create Redis client using redis.NewClient()
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	// 2. Test connection with Ping()
	rdbing := rdb.Ping(context.Background())
	if rdbing.Err() != nil {
		log.Fatalln(rdbing.Err())
		return nil, rdbing.Err()
	}
	// 3. Return RedisCache instance
	return &RedisCache{
		client: rdb,
	}, nil
}

// SaveFileMetadata saves file metadata to Redis
func (r *RedisCache) SaveFileMetadata(ctx context.Context, fileID string, metadata *FileMetadata, expiration time.Duration) error {
	// 1. Marshal metadata to JSON
	data, err := json.Marshal(metadata)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	// 2. Use client.Set() to save with key "file:{fileID}"
	return r.client.Set(ctx, "file:"+fileID, data, expiration).Err()
}

// GetFileMetadata retrieves file metadata from Redis
func (r *RedisCache) GetFileMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	// 1. Use client.Get() to retrieve with key "file:{fileID}"
	result, err := r.client.Get(ctx, "file:"+fileID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, err // Key does not exist
		}
		log.Fatalln(err)
		return nil, err
	}
	// 2. Unmarshal JSON to FileMetadata struct
	var metadata FileMetadata
	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return &metadata, nil
}

// DeleteFileMetadata deletes file metadata
func (r *RedisCache) DeleteFileMetadata(ctx context.Context, fileID string) error {
	// Use client.Del() to delete the key
	return r.client.Del(ctx, "file:"+fileID).Err()
}

// AddFileToUserIndex adds file to user's file list
func (r *RedisCache) AddFileToUserIndex(ctx context.Context, userID, fileID string) error {
	// 1. Use client.SAdd() to add to set "user:{userID}:files"
	// 2. This allows listing all files for a user
	return r.client.SAdd(ctx, "user:"+userID+":files", fileID).Err()
}

// GetUserFiles retrieves all file IDs for a user
func (r *RedisCache) GetUserFiles(ctx context.Context, userID string) ([]string, error) {
	// 1. Use client.SMembers() to get all files in set "user:{userID}:files"
	// 2. Return slice of file IDs
	return r.client.SMembers(ctx, "user:"+userID+":files").Result()
}

// SaveSession saves a user session
func (r *RedisCache) SaveSession(ctx context.Context, token, userID string, expiration time.Duration) error {
	// 1. Use client.Set() with key "session:{token}"
	// 2. Value should be userID
	// 3. Set expiration (session timeout)
	return r.client.Set(ctx, "session:"+token, userID, expiration).Err()
}

// GetSession retrieves user ID from session token
func (r *RedisCache) GetSession(ctx context.Context, token string) (string, error) {
	// Use client.Get() with key "session:{token}"
	return r.client.Get(ctx, "session:"+token).Result()
}

// DeleteSession deletes a session (logout)
func (r *RedisCache) DeleteSession(ctx context.Context, token string) error {
	// Use client.Del() to delete session
	return r.client.Del(ctx, "session:"+token).Err()
}
