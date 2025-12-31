package api

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sachinthra/file-locker/backend/internal/constants"
	"github.com/sachinthra/file-locker/backend/internal/crypto"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type StreamHandler struct {
	minioStorage *storage.MinIOStorage
	redisCache   *storage.RedisCache
	pgStore      *storage.PostgresStore
}

func NewStreamHandler(minioStorage *storage.MinIOStorage, redisCache *storage.RedisCache, pgStore *storage.PostgresStore) *StreamHandler {
	return &StreamHandler{
		minioStorage: minioStorage,
		redisCache:   redisCache,
		pgStore:      pgStore,
	}
}

func (h *StreamHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	// 1. Get fileID from URL
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		respondError(w, http.StatusBadRequest, "File ID required")
		return
	}

	// 2. Get userID from context (Security Check)
	userID, ok := r.Context().Value(constants.UserIDKey).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// 3. Get metadata from PostgreSQL
	metadata, err := h.pgStore.GetFileMetadata(r.Context(), fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	// 4. Verify Ownership
	if metadata.UserID != userID {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// 5. Check Expiration
	if metadata.ExpiresAt != nil && metadata.ExpiresAt.Before(time.Now()) {
		respondError(w, http.StatusGone, "File has expired")
		return
	}

	// 6. Decode the Master Encryption Key
	keyBytes, err := base64.StdEncoding.DecodeString(metadata.EncryptionKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode encryption key")
		return
	}

	// 7. Handle Range Request (Seeking) vs Full Request
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		h.handleRangeRequest(w, r, metadata, keyBytes, rangeHeader)
	} else {
		h.handleFullStream(w, r, metadata, keyBytes)
	}
}

// handleFullStream decrypts the entire file from start to finish
func (h *StreamHandler) handleFullStream(w http.ResponseWriter, r *http.Request, metadata *storage.FileMetadata, keyBytes []byte) {
	// Fetch entire encrypted stream from MinIO
	encryptedStream, err := h.minioStorage.GetFile(r.Context(), metadata.MinIOPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve file")
		return
	}
	defer encryptedStream.Close()

	// Use our existing helper which reads the IV from the first 16 bytes automatically
	decryptedStream, err := crypto.DecryptStream(encryptedStream, keyBytes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decrypt file")
		return
	}

	// Standard Headers
	w.Header().Set("Content-Type", metadata.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", metadata.Size))
	w.Header().Set("Accept-Ranges", "bytes") // Tells browser we support seeking
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", metadata.FileName))
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	// Stream data
	if _, err := io.Copy(w, decryptedStream); err != nil {
		// Connection likely closed by client
		return
	}
}

// handleRangeRequest handles seeking by calculating the correct AES Counter offset
func (h *StreamHandler) handleRangeRequest(w http.ResponseWriter, r *http.Request, metadata *storage.FileMetadata, keyBytes []byte, rangeHeader string) {
	// 1. Parse Range Header: "bytes=1000-2000"
	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	rangeParts := strings.Split(rangeStr, "-")

	start, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid range start")
		return
	}

	var end int64
	if len(rangeParts) > 1 && rangeParts[1] != "" {
		end, err = strconv.ParseInt(rangeParts[1], 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid range end")
			return
		}
	} else {
		end = metadata.Size - 1 // Default to end of file
	}

	if start > end || start >= metadata.Size {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", metadata.Size))
		respondError(w, http.StatusRequestedRangeNotSatisfiable, "Invalid range")
		return
	}

	// 2. Calculate AES Block Alignment
	// AES-GCM/CTR works on 16-byte blocks. We need to find which block our 'start' byte lives in.
	const blockSize = 16
	const ivSize = 16

	blockNumber := uint64(start / blockSize) // Which block index (0, 1, 2...)
	offsetInBlock := start % blockSize       // How far into that block (0-15)

	// 3. Fetch the Original IV (First 16 bytes of file)
	// We need this to calculate the specific counter for our block.
	ivStream, err := h.minioStorage.GetFileRange(r.Context(), metadata.MinIOPath, 0, int64(ivSize-1))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve IV")
		return
	}
	iv := make([]byte, ivSize)
	if _, err := io.ReadFull(ivStream, iv); err != nil {
		ivStream.Close()
		respondError(w, http.StatusInternalServerError, "Failed to read IV")
		return
	}
	ivStream.Close()

	// 4. Calculate the Counter for this specific block
	// CTR mode works by encrypting (IV + Counter). We manually add blockNumber to IV.
	currentIV := addCounter(iv, blockNumber)

	// 5. Fetch Encrypted Data from MinIO
	// We start fetching from the beginning of the block to ensure decryption alignment.
	// MinIO Offset = IV Size + (Block Index * 16)
	fetchStart := int64(ivSize) + (int64(blockNumber) * blockSize)
	fetchEnd := int64(ivSize) + end

	encryptedStream, err := h.minioStorage.GetFileRange(r.Context(), metadata.MinIOPath, fetchStart, fetchEnd)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve file range")
		return
	}
	defer encryptedStream.Close()

	// 6. Initialize Cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create cipher")
		return
	}

	// Create CTR stream starting at our calculated IV
	stream := cipher.NewCTR(block, currentIV)

	// 7. Set Response Headers
	contentLength := end - start + 1
	w.Header().Set("Content-Type", metadata.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, metadata.Size))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)

	// 8. Decrypt and Stream
	// Buffer size: 32KB
	buf := make([]byte, 32*1024)

	// We might need to discard bytes if 'start' wasn't exactly on a block boundary
	firstChunk := true

	for {
		n, err := encryptedStream.Read(buf)
		if n > 0 {
			// Decrypt in place
			stream.XORKeyStream(buf[:n], buf[:n])

			writeBuf := buf[:n]

			// If this is the first chunk, trim the leading bytes we fetched for alignment but user didn't ask for
			if firstChunk {
				if int64(n) > offsetInBlock {
					writeBuf = buf[offsetInBlock:n]
				} else {
					// Edge case: chunk is smaller than offset (unlikely with 32KB buf)
					offsetInBlock -= int64(n)
					continue
				}
				firstChunk = false
			}

			if _, wErr := w.Write(writeBuf); wErr != nil {
				return // Client disconnected
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			// Stream broken mid-way
			return
		}
	}
}

// addCounter increments an AES-CTR 16-byte counter by a specific value (Big Endian addition)
func addCounter(iv []byte, delta uint64) []byte {
	// Create a copy so we don't modify the original IV
	newIV := make([]byte, len(iv))
	copy(newIV, iv)

	// Add delta to the byte array (treating it as a big-endian integer)
	// We iterate backwards through the byte slice
	for i := len(newIV) - 1; i >= 0; i-- {
		sum := uint64(newIV[i]) + (delta & 0xFF)
		newIV[i] = byte(sum)

		// Shift delta for next byte and handle carry
		delta >>= 8
		if sum > 255 {
			delta++
		}

		// Optimization: if no more delta to add, stop
		if delta == 0 {
			break
		}
	}
	return newIV
}
