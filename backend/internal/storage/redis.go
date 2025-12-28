package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache handles ephemeral data: sessions, rate limiting, and caching
// Permanent data (users, files) moved to PostgreSQL
type RedisCache struct {
	client *redis.Client
}

// FileMetadata is now primarily stored in PostgreSQL
// This struct is kept here for compatibility and caching purposes
type FileMetadata struct {
	FileID        string     `json:"file_id"`
	UserID        string     `json:"user_id"`
	FileName      string     `json:"file_name"`
	DisplayName   string     `json:"display_name,omitempty"`
	Description   string     `json:"description,omitempty"`
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

// =====================================================
// RATE LIMITING (EPHEMERAL - STAYS IN REDIS)
// =====================================================

// IncrRateLimit increments the rate limit counter for a user in a time window
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

// =====================================================
// SESSION MANAGEMENT (EPHEMERAL - STAYS IN REDIS)
// =====================================================

// SaveSession stores a JWT session token
func (r *RedisCache) SaveSession(ctx context.Context, token, userID string, expiration time.Duration) error {
	return r.client.Set(ctx, "session:"+token, userID, expiration).Err()
}

// GetSession retrieves the userID for a given session token
func (r *RedisCache) GetSession(ctx context.Context, token string) (string, error) {
	return r.client.Get(ctx, "session:"+token).Result()
}

// DeleteSession removes a session token
func (r *RedisCache) DeleteSession(ctx context.Context, token string) error {
	return r.client.Del(ctx, "session:"+token).Err()
}
