package storage

import (
	"context"
	"testing"
	"time"
)

// Note: These tests require a running Redis instance
// Skip tests if Redis is not available

func setupRedisTest(t *testing.T) *RedisCache {
	redis, err := NewRedisCache("localhost:6379", "", 0)
	if err != nil {
		t.Skip("Redis not available, skipping tests")
		return nil
	}
	return redis
}

func TestNewRedisCache(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	if redis.client == nil {
		t.Error("Expected client to be initialized")
	}
}

func TestSaveAndGetFileMetadata(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	fileID := "test-file-123"

	metadata := &FileMetadata{
		FileID:        fileID,
		UserID:        "user-456",
		FileName:      "test.txt",
		MimeType:      "text/plain",
		Size:          1024,
		EncryptedSize: 1056,
		MinIOPath:     "/bucket/test.txt",
		EncryptionKey: "encrypted-key-data",
		CreatedAt:     time.Now(),
		Tags:          []string{"test", "sample"},
		DownloadCount: 0,
	}

	// Save metadata
	err := redis.SaveFileMetadata(ctx, fileID, metadata, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Get metadata
	retrieved, err := redis.GetFileMetadata(ctx, fileID)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	// Verify
	if retrieved.FileID != metadata.FileID {
		t.Errorf("Expected FileID %s, got %s", metadata.FileID, retrieved.FileID)
	}
	if retrieved.UserID != metadata.UserID {
		t.Errorf("Expected UserID %s, got %s", metadata.UserID, retrieved.UserID)
	}
	if retrieved.FileName != metadata.FileName {
		t.Errorf("Expected FileName %s, got %s", metadata.FileName, retrieved.FileName)
	}

	// Cleanup
	redis.DeleteFileMetadata(ctx, fileID)
}

func TestGetFileMetadata_NotFound(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	_, err := redis.GetFileMetadata(ctx, "nonexistent-file")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestDeleteFileMetadata(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	fileID := "test-delete-123"

	metadata := &FileMetadata{
		FileID:   fileID,
		UserID:   "user-123",
		FileName: "delete-test.txt",
	}

	// Save
	err := redis.SaveFileMetadata(ctx, fileID, metadata, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Delete
	err = redis.DeleteFileMetadata(ctx, fileID)
	if err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}

	// Verify deleted
	_, err = redis.GetFileMetadata(ctx, fileID)
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func TestAddFileToUserIndex(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	userID := "user-789"
	fileID := "file-101"

	// Add file to user index
	err := redis.AddFileToUserIndex(ctx, userID, fileID)
	if err != nil {
		t.Fatalf("Failed to add file to user index: %v", err)
	}

	// Get user files
	files, err := redis.GetUserFiles(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get user files: %v", err)
	}

	// Verify file is in list
	found := false
	for _, f := range files {
		if f == fileID {
			found = true
			break
		}
	}
	if !found {
		t.Error("File not found in user index")
	}

	// Cleanup
	redis.client.Del(ctx, "user:"+userID+":files")
}

func TestGetUserFiles_Empty(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	files, err := redis.GetUserFiles(ctx, "nonexistent-user")
	if err != nil {
		t.Fatalf("Expected no error for empty user files, got %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty file list, got %d files", len(files))
	}
}

func TestSaveAndGetSession(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	token := "test-token-abc123"
	userID := "user-999"

	// Save session
	err := redis.SaveSession(ctx, token, userID, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Get session
	retrievedUserID, err := redis.GetSession(ctx, token)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, retrievedUserID)
	}

	// Cleanup
	redis.DeleteSession(ctx, token)
}

func TestGetSession_NotFound(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	_, err := redis.GetSession(ctx, "nonexistent-token")
	if err == nil {
		t.Error("Expected error for nonexistent session, got nil")
	}
}

func TestDeleteSession(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	token := "test-delete-token"
	userID := "user-111"

	// Save session
	err := redis.SaveSession(ctx, token, userID, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Delete session
	err = redis.DeleteSession(ctx, token)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify deleted
	_, err = redis.GetSession(ctx, token)
	if err == nil {
		t.Error("Expected error after session deletion, got nil")
	}
}

func TestExists(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	key := "test-exists-key"

	// Key should not exist initially
	exists, err := redis.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected key to not exist initially")
	}

	// Set key
	err = redis.Set(ctx, key, "value", 1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Key should exist now
	exists, err = redis.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist after setting")
	}

	// Cleanup
	redis.client.Del(ctx, key)
}

func TestIncr(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	key := "test-incr-key"

	// First increment
	count, err := redis.IncrRateLimit(ctx, key, 0)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Second increment
	count, err = redis.IncrRateLimit(ctx, key, 0)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Third increment
	count, err = redis.IncrRateLimit(ctx, key, 0)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Cleanup
	redis.client.Del(ctx, key)
}

func TestSet(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	key := "test-set-key"
	value := "test-value"

	// Set key with expiration
	err := redis.Set(ctx, key, value, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Verify key exists
	exists, err := redis.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist after setting")
	}

	// Get value
	result, err := redis.client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if result != value {
		t.Errorf("Expected value %s, got %s", value, result)
	}

	// Cleanup
	redis.client.Del(ctx, key)
}

func TestFileMetadata_WithExpiration(t *testing.T) {
	redis := setupRedisTest(t)
	if redis == nil {
		return
	}

	ctx := context.Background()
	fileID := "expiring-file"

	metadata := &FileMetadata{
		FileID:   fileID,
		UserID:   "user-exp",
		FileName: "expiring.txt",
	}

	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	metadata.ExpiresAt = &expiresAt

	// Save with short expiration for testing
	err := redis.SaveFileMetadata(ctx, fileID, metadata, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Should exist immediately
	retrieved, err := redis.GetFileMetadata(ctx, fileID)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if retrieved.ExpiresAt == nil {
		t.Error("Expected ExpiresAt to be set")
	}

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Should not exist after expiration
	_, err = redis.GetFileMetadata(ctx, fileID)
	if err == nil {
		t.Error("Expected error after expiration, got nil")
	}
}
