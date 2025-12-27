package api

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileLifecycle(t *testing.T) {
	// Setup: Create User
	userID, token := createTestUser(t, "flow_user")
	t.Log(token)
	ctx := context.WithValue(context.Background(), "userID", userID)

	// ---------------------------------------------------------
	// 1. TEST UPLOAD
	// ---------------------------------------------------------
	fileContent := []byte("This is a test file content for streaming and encryption verification.")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test_stream.txt")
	part.Write(fileContent)
	writer.Close()

	reqUpload := httptest.NewRequest("POST", "/upload", body)
	reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
	// Inject Context (Simulate Auth Middleware)
	reqUpload = reqUpload.WithContext(ctx)

	wUpload := httptest.NewRecorder()
	testUpload.HandleUpload(wUpload, reqUpload)

	require.Equal(t, http.StatusCreated, wUpload.Code)

	var uploadResp UploadResponse
	json.NewDecoder(wUpload.Body).Decode(&uploadResp)
	fileID := uploadResp.FileID
	require.NotEmpty(t, fileID)

	// ---------------------------------------------------------
	// 2. TEST STREAMING (Range Request)
	// ---------------------------------------------------------
	// Let's try to read bytes 10-20: "test file c"
	reqStream := httptest.NewRequest("GET", "/stream/"+fileID, nil)
	reqStream = reqStream.WithContext(ctx)

	// Add Range Header
	reqStream.Header.Set("Range", "bytes=10-20")

	// Setup Chi Context for URL Param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fileID)
	reqStream = reqStream.WithContext(context.WithValue(reqStream.Context(), chi.RouteCtxKey, rctx))

	wStream := httptest.NewRecorder()
	testStream.HandleStream(wStream, reqStream)

	require.Equal(t, http.StatusPartialContent, wStream.Code)

	// Verify Content
	expectedChunk := fileContent[10:21] // Go slice is exclusive at end, Range is inclusive
	assert.Equal(t, expectedChunk, wStream.Body.Bytes(), "Decrypted range does not match plaintext range!")

	// ---------------------------------------------------------
	// 3. TEST LIST FILES
	// ---------------------------------------------------------
	reqList := httptest.NewRequest("GET", "/files", nil)
	reqList = reqList.WithContext(ctx)
	wList := httptest.NewRecorder()

	testFiles.HandleListFiles(wList, reqList)
	assert.Equal(t, http.StatusOK, wList.Code)
	assert.Contains(t, wList.Body.String(), "test_stream.txt")

	// ---------------------------------------------------------
	// 4. TEST DELETE
	// ---------------------------------------------------------
	reqDel := httptest.NewRequest("DELETE", "/files?id="+fileID, nil)
	reqDel = reqDel.WithContext(ctx)
	wDel := httptest.NewRecorder()

	testFiles.HandleDeleteFile(wDel, reqDel)
	assert.Equal(t, http.StatusOK, wDel.Code)

	// Verify it's gone from Redis
	_, err := testRedis.GetFileMetadata(context.Background(), fileID)
	assert.Error(t, err)
}
