package crypto

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)
	assert.Len(t, key, 32)

	key2, err := GenerateKey()
	require.NoError(t, err)
	assert.NotEqual(t, key, key2, "Keys should be random")
}

func TestEncryptDecryptBytes(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("Hello, World! This is a secret message.")

	// Test Encrypt
	ciphertext, err := EncryptBytes(plaintext, key)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	// Test Decrypt
	decrypted, err := DecryptBytes(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestStreamEncryption(t *testing.T) {
	key, _ := GenerateKey()
	inputData := make([]byte, 1024*1024) // 1MB dummy data
	// Fill with some pattern
	for i := range inputData {
		inputData[i] = byte(i % 255)
	}

	// 1. Encrypt
	reader := bytes.NewReader(inputData)
	encryptedReader, err := EncryptStream(reader, key)
	require.NoError(t, err)

	// Read all encrypted data
	encryptedData, err := io.ReadAll(encryptedReader)
	require.NoError(t, err)

	// Ensure it's not the same as input
	assert.NotEqual(t, inputData, encryptedData)
	// Encrypted size should be Input + 16 bytes (IV)
	assert.Equal(t, len(inputData)+16, len(encryptedData))

	// 2. Decrypt
	cipherReader := bytes.NewReader(encryptedData)
	decryptedReader, err := DecryptStream(cipherReader, key)
	require.NoError(t, err)

	decryptedData, err := io.ReadAll(decryptedReader)
	require.NoError(t, err)

	// 3. Verify
	assert.Equal(t, inputData, decryptedData)
}

func TestEncryptBytesInvalidKey(t *testing.T) {
	shortKey := []byte("short")
	_, err := EncryptBytes([]byte("data"), shortKey)
	assert.Error(t, err)
}
