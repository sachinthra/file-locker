package api

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sachinthra/file-locker/backend/internal/auth"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

// Global test variables
var (
	testRedis  *storage.RedisCache
	testMinIO  *storage.MinIOStorage
	testJWT    *auth.JWTService
	testAuth   *AuthHandler
	testUpload *UploadHandler
	testFiles  *FilesHandler
	testStream *StreamHandler
	testDown   *DownloadHandler
)

func TestMain(m *testing.M) {
	// 1. Setup Dependencies (Connect to Docker containers)
	var err error

	// Redis (localhost:6379 from docker-compose)
	testRedis, err = storage.NewRedisCache("localhost:6379", "", 0)
	if err != nil {
		log.Printf("Skipping integration tests: Redis not running (%v)", err)
		return // Exit if no DB
	}

	// MinIO (localhost:9012 from docker-compose mapping)
	testMinIO, err = storage.NewMinIOStorage(
		"localhost:9012",
		"minioadmin", // Default from docker-compose
		"minioadmin",
		"test-bucket", // Separate bucket for testing
		false,
		"us-east-1",
	)
	if err != nil {
		log.Printf("Skipping integration tests: MinIO not running (%v)", err)
		return
	}

	testJWT = auth.NewJWTService("test-secret", 3600)

	// 2. Initialize Handlers
	testAuth = NewAuthHandler(testJWT, testRedis)
	testUpload = NewUploadHandler(testMinIO, testRedis)
	testFiles = NewFilesHandler(testRedis, testMinIO)
	testStream = NewStreamHandler(testMinIO, testRedis)
	testDown = NewDownloadHandler(testMinIO, testRedis)

	// 3. Run Tests
	code := m.Run()

	// 4. Cleanup (Optional: Flush Redis)
	// testRedis.Client.FlushDB(context.Background())

	os.Exit(code)
}

// Helper to create a dummy user and get token
func createTestUser(t *testing.T, username string) (string, string) {
	ctx := context.Background()
	userID := "user-" + username
	// Create user manually in Redis
	err := testRedis.SaveUser(ctx, username, userID, "$2a$10$dummyhash", "test@example.com", 0)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	token, _ := testJWT.GenerateToken(userID)
	// Create session
	testRedis.SaveSession(ctx, token, userID, time.Hour)

	return userID, token
}
