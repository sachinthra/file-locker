package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

type FileMetadata struct {
	FileID        string     `json:"file_id"`
	UserID        string     `json:"user_id"`
	FileName      string     `json:"file_name"`
	MimeType      string     `json:"mime_type"`
	Size          int64      `json:"size"`
	EncryptedSize int64      `json:"encrypted_size"`
	MinIOPath     string     `json:"minio_path"`
	EncryptionKey string     `json:"encryption_key"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	DownloadCount int        `json:"download_count"`
}

type UserData struct {
	UserID       string `json:"user_id"`
	PasswordHash string `json:"password_hash"`
	Email        string `json:"email"`
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: rdb}, nil
}

// Basic key-value operations

func (r *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return result > 0, nil
}

// File metadata management functions

func (r *RedisCache) SaveFileMetadata(ctx context.Context, fileID string, metadata *FileMetadata, expiration time.Duration) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	return r.client.Set(ctx, "file:"+fileID, data, expiration).Err()
}

func (r *RedisCache) GetFileMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	result, err := r.client.Get(ctx, "file:"+fileID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("file not found: %s", fileID)
		}
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	var metadata FileMetadata
	if err := json.Unmarshal([]byte(result), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	return &metadata, nil
}

// Rate limiting functions

func (r *RedisCache) IncrRateLimit(ctx context.Context, userID string, currentWindow int64) (int64, error) {
	rateLimitKey := fmt.Sprintf("ratelimit:%s:%d", userID, currentWindow)
	result, err := r.client.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key: %w", err)
	}
	return result, nil
}

func (r *RedisCache) SetRateLimit(ctx context.Context, userID string, currentWindow int64, value string, expiration time.Duration) error {
	rateLimitKey := fmt.Sprintf("ratelimit:%s:%d", userID, currentWindow)
	return r.client.Set(ctx, rateLimitKey, value, expiration).Err()
}

// User management functions

func (r *RedisCache) UserExists(ctx context.Context, username string) (bool, error) {
	userKey := "user:" + username
	result, err := r.client.Exists(ctx, userKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return result > 0, nil
}

func (r *RedisCache) SaveUser(ctx context.Context, username, userID, hashedPassword, email string, expiration time.Duration) error {
	data := UserData{
		UserID:       userID,
		PasswordHash: hashedPassword,
		Email:        email,
	}
	jsonData, _ := json.Marshal(data)
	return r.client.Set(ctx, "user:"+username, jsonData, expiration).Err()
}

func (r *RedisCache) GetUser(ctx context.Context, username string) (*UserData, error) {
	val, err := r.client.Get(ctx, "user:"+username).Result()
	if err != nil {
		return nil, err
	}

	var data UserData
	json.Unmarshal([]byte(val), &data)
	return &data, nil
}

// File metadata management functions

func (r *RedisCache) DeleteFileMetadata(ctx context.Context, fileID string) error {
	return r.client.Del(ctx, "file:"+fileID).Err()
}

func (r *RedisCache) AddFileToUserIndex(ctx context.Context, userID, fileID string) error {
	return r.client.SAdd(ctx, "user:"+userID+":files", fileID).Err()
}

func (r *RedisCache) GetUserFiles(ctx context.Context, userID string) ([]string, error) {
	return r.client.SMembers(ctx, "user:"+userID+":files").Result()
}

func (r *RedisCache) RemoveFileFromUserIndex(ctx context.Context, userID, fileID string) error {
	return r.client.SRem(ctx, "user:"+userID+":files", fileID).Err()
}

// Session management functions

func (r *RedisCache) SaveSession(ctx context.Context, token, userID string, expiration time.Duration) error {
	return r.client.Set(ctx, "session:"+token, userID, expiration).Err()
}

func (r *RedisCache) GetSession(ctx context.Context, token string) (string, error) {
	return r.client.Get(ctx, "session:"+token).Result()
}

func (r *RedisCache) DeleteSession(ctx context.Context, token string) error {
	return r.client.Del(ctx, "session:"+token).Err()
}
