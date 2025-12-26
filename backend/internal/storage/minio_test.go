package storage

import (
	"bytes"
	"context"
	"io"
	"log"
	"testing"

	"github.com/sachinthra/file-locker/backend/internal/config"
)

func getConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}

func setupMinIOStorage(t *testing.T) *MinIOStorage {
	cfg := getConfig()
	m, err := NewMinIOStorage(
		cfg.Storage.MinIO.Endpoint,
		cfg.Storage.MinIO.AccessKey,
		cfg.Storage.MinIO.SecretKey,
		cfg.Storage.MinIO.Bucket,
		cfg.Storage.MinIO.UseSSL,
		cfg.Storage.MinIO.Region,
	)
	if err != nil {
		t.Fatalf("Failed to create MinIO storage: %v", err)
	}
	return m
}

func TestNewMinIOStorage(t *testing.T) {
	m := setupMinIOStorage(t)
	if m == nil {
		t.Fatal("MinIO storage should not be nil")
	}
	if m.client == nil {
		t.Fatal("MinIO client should not be nil")
	}
	t.Logf("MinIO storage created successfully")
}

func TestSaveFile(t *testing.T) {
	m := setupMinIOStorage(t)
	ctx := context.Background()

	// Create test data
	testData := []byte("Hello, this is a test file for MinIO storage!")
	testReader := bytes.NewReader(testData)
	objectName := "test-save-file.txt"

	// Save file
	err := m.SaveFile(ctx, objectName, testReader, int64(len(testData)), "text/plain")
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	t.Logf("File saved successfully: %s", objectName)

	// Cleanup
	defer m.DeleteFile(ctx, objectName)
}

func TestGetFile(t *testing.T) {
	m := setupMinIOStorage(t)
	ctx := context.Background()

	// First, upload a test file
	testData := []byte("Test data for GetFile function")
	objectName := "test-get-file.txt"
	err := m.SaveFile(ctx, objectName, bytes.NewReader(testData), int64(len(testData)), "text/plain")
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}
	defer m.DeleteFile(ctx, objectName)

	// Now try to get the file
	reader, err := m.GetFile(ctx, objectName)
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}
	defer reader.Close()

	// Read the content
	downloadedData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	// Verify content matches
	if !bytes.Equal(testData, downloadedData) {
		t.Errorf("Downloaded data doesn't match. Expected: %s, Got: %s", testData, downloadedData)
	}

	t.Logf("File retrieved successfully and content matches")
}

func TestGetFileRange(t *testing.T) {
	m := setupMinIOStorage(t)
	ctx := context.Background()

	// Create a larger test file
	testData := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	objectName := "test-range-file.txt"
	err := m.SaveFile(ctx, objectName, bytes.NewReader(testData), int64(len(testData)), "text/plain")
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}
	defer m.DeleteFile(ctx, objectName)

	// Test getting a range (bytes 10-19, which should be "ABCDEFGHIJ")
	reader, err := m.GetFileRange(ctx, objectName, 10, 19)
	if err != nil {
		t.Fatalf("Failed to get file range: %v", err)
	}
	defer reader.Close()

	// Read the range content
	rangeData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read range data: %v", err)
	}

	// Verify we got the correct range
	expectedRange := testData[10:20] // bytes 10-19 inclusive
	if !bytes.Equal(expectedRange, rangeData) {
		t.Errorf("Range data doesn't match. Expected: %s, Got: %s", expectedRange, rangeData)
	}

	t.Logf("File range retrieved successfully: %s", rangeData)
}

func TestGetFileInfo(t *testing.T) {
	m := setupMinIOStorage(t)
	ctx := context.Background()

	// Upload a test file
	testData := []byte("Test data for GetFileInfo")
	objectName := "test-info-file.txt"
	contentType := "text/plain"
	err := m.SaveFile(ctx, objectName, bytes.NewReader(testData), int64(len(testData)), contentType)
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}
	defer m.DeleteFile(ctx, objectName)

	// Get file info
	info, err := m.GetFileInfo(ctx, objectName)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	// Verify info
	if info.Size != int64(len(testData)) {
		t.Errorf("File size doesn't match. Expected: %d, Got: %d", len(testData), info.Size)
	}

	if info.ContentType != contentType {
		t.Errorf("Content type doesn't match. Expected: %s, Got: %s", contentType, info.ContentType)
	}

	if info.Key != objectName {
		t.Errorf("Object name doesn't match. Expected: %s, Got: %s", objectName, info.Key)
	}

	t.Logf("File info retrieved successfully: Size=%d, ContentType=%s, LastModified=%s",
		info.Size, info.ContentType, info.LastModified)
}

func TestDeleteFile(t *testing.T) {
	m := setupMinIOStorage(t)
	ctx := context.Background()

	// Upload a test file
	testData := []byte("Test data for delete")
	objectName := "test-delete-file.txt"
	err := m.SaveFile(ctx, objectName, bytes.NewReader(testData), int64(len(testData)), "text/plain")
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	// Verify file exists
	_, err = m.GetFileInfo(ctx, objectName)
	if err != nil {
		t.Fatalf("File should exist before deletion: %v", err)
	}

	// Delete the file
	err = m.DeleteFile(ctx, objectName)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file no longer exists
	_, err = m.GetFileInfo(ctx, objectName)
	if err == nil {
		t.Error("File should not exist after deletion")
	}

	t.Logf("File deleted successfully")
}

func TestFullWorkflow(t *testing.T) {
	m := setupMinIOStorage(t)
	ctx := context.Background()

	objectName := "test-workflow-file.txt"
	testData := []byte("This is a complete workflow test!")

	// Step 1: Save file
	t.Log("Step 1: Saving file...")
	err := m.SaveFile(ctx, objectName, bytes.NewReader(testData), int64(len(testData)), "text/plain")
	if err != nil {
		t.Fatalf("Step 1 failed - Save: %v", err)
	}

	// Step 2: Get file info
	t.Log("Step 2: Getting file info...")
	info, err := m.GetFileInfo(ctx, objectName)
	if err != nil {
		t.Fatalf("Step 2 failed - GetInfo: %v", err)
	}
	t.Logf("File info: Size=%d, Type=%s", info.Size, info.ContentType)

	// Step 3: Get full file
	t.Log("Step 3: Downloading full file...")
	reader, err := m.GetFile(ctx, objectName)
	if err != nil {
		t.Fatalf("Step 3 failed - GetFile: %v", err)
	}
	fullData, _ := io.ReadAll(reader)
	reader.Close()
	if !bytes.Equal(testData, fullData) {
		t.Error("Step 3 failed - Content mismatch")
	}

	// Step 4: Get partial range
	t.Log("Step 4: Downloading partial range...")
	rangeReader, err := m.GetFileRange(ctx, objectName, 0, 9)
	if err != nil {
		t.Fatalf("Step 4 failed - GetFileRange: %v", err)
	}
	rangeData, _ := io.ReadAll(rangeReader)
	rangeReader.Close()
	expectedRange := testData[0:10]
	if !bytes.Equal(expectedRange, rangeData) {
		t.Errorf("Step 4 failed - Range mismatch. Expected: %s, Got: %s", expectedRange, rangeData)
	}

	// Step 5: Delete file
	t.Log("Step 5: Deleting file...")
	err = m.DeleteFile(ctx, objectName)
	if err != nil {
		t.Fatalf("Step 5 failed - Delete: %v", err)
	}

	// Step 6: Verify deletion
	t.Log("Step 6: Verifying deletion...")
	_, err = m.GetFileInfo(ctx, objectName)
	if err == nil {
		t.Error("Step 6 failed - File still exists after deletion")
	}

	t.Log("✅ Full workflow test completed successfully!")
}
