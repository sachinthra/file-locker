package crypto

import (
	"bytes"
	"io"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Generate another key and ensure they're different
	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bytes.Equal(key, key2) {
		t.Error("Expected different keys, got identical keys")
	}
}

func TestEncryptDecryptStream(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	originalData := []byte("This is a test message for stream encryption")
	plaintext := bytes.NewReader(originalData)

	// Encrypt
	encrypted, err := EncryptStream(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Read encrypted data
	encryptedData, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatalf("Failed to read encrypted data: %v", err)
	}

	// Encrypted data should be different from original
	if bytes.Equal(encryptedData, originalData) {
		t.Error("Encrypted data should differ from original")
	}

	// Decrypt
	ciphertext := bytes.NewReader(encryptedData)
	decrypted, err := DecryptStream(ciphertext, key)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Read decrypted data
	decryptedData, err := io.ReadAll(decrypted)
	if err != nil {
		t.Fatalf("Failed to read decrypted data: %v", err)
	}

	// Decrypted should match original
	if !bytes.Equal(decryptedData, originalData) {
		t.Errorf("Decrypted data doesn't match original.\nExpected: %s\nGot: %s", originalData, decryptedData)
	}
}

func TestEncryptStream_LargeData(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create 1MB of data
	originalData := bytes.Repeat([]byte("A"), 1024*1024)
	plaintext := bytes.NewReader(originalData)

	encrypted, err := EncryptStream(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to encrypt large data: %v", err)
	}

	encryptedData, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatalf("Failed to read encrypted data: %v", err)
	}

	ciphertext := bytes.NewReader(encryptedData)
	decrypted, err := DecryptStream(ciphertext, key)
	if err != nil {
		t.Fatalf("Failed to decrypt large data: %v", err)
	}

	decryptedData, err := io.ReadAll(decrypted)
	if err != nil {
		t.Fatalf("Failed to read decrypted data: %v", err)
	}

	if !bytes.Equal(decryptedData, originalData) {
		t.Error("Decrypted large data doesn't match original")
	}
}

func TestEncryptStream_InvalidKeyLength(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
	}{
		{"too short", 8},
		{"odd length", 17},
		{"too long", 48},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLength)
			plaintext := bytes.NewReader([]byte("test"))

			_, err := EncryptStream(plaintext, key)
			if err == nil {
				t.Errorf("Expected error for key length %d, got nil", tt.keyLength)
			}
		})
	}
}

func TestDecryptStream_InvalidKeyLength(t *testing.T) {
	invalidKey := make([]byte, 8)
	ciphertext := bytes.NewReader([]byte("test"))

	_, err := DecryptStream(ciphertext, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key length, got nil")
	}
}

func TestEncryptDecryptBytes(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	originalData := []byte("Secret metadata content")

	// Encrypt
	encrypted, err := EncryptBytes(originalData, key)
	if err != nil {
		t.Fatalf("Failed to encrypt bytes: %v", err)
	}

	// Encrypted should be different
	if bytes.Equal(encrypted, originalData) {
		t.Error("Encrypted bytes should differ from original")
	}

	// Encrypted should be longer (nonce + ciphertext + tag)
	if len(encrypted) <= len(originalData) {
		t.Error("Encrypted data should be longer than original")
	}

	// Decrypt
	decrypted, err := DecryptBytes(encrypted, key)
	if err != nil {
		t.Fatalf("Failed to decrypt bytes: %v", err)
	}

	// Should match original
	if !bytes.Equal(decrypted, originalData) {
		t.Errorf("Decrypted bytes don't match original.\nExpected: %s\nGot: %s", originalData, decrypted)
	}
}

func TestEncryptBytes_EmptyData(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encrypted, err := EncryptBytes([]byte{}, key)
	if err != nil {
		t.Fatalf("Failed to encrypt empty data: %v", err)
	}

	decrypted, err := DecryptBytes(encrypted, key)
	if err != nil {
		t.Fatalf("Failed to decrypt empty data: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("Expected empty decrypted data, got %d bytes", len(decrypted))
	}
}

func TestEncryptBytes_InvalidKeyLength(t *testing.T) {
	invalidKey := make([]byte, 10)
	data := []byte("test")

	_, err := EncryptBytes(data, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key length, got nil")
	}
}

func TestDecryptBytes_InvalidKeyLength(t *testing.T) {
	invalidKey := make([]byte, 10)
	data := []byte("test")

	_, err := DecryptBytes(data, invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key length, got nil")
	}
}

func TestDecryptBytes_TamperedData(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	originalData := []byte("Important secret data")
	encrypted, err := EncryptBytes(originalData, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Tamper with the encrypted data
	if len(encrypted) > 10 {
		encrypted[10] ^= 0xFF
	}

	// Decryption should fail
	_, err = DecryptBytes(encrypted, key)
	if err == nil {
		t.Error("Expected error for tampered data, got nil")
	}
}

func TestDecryptBytes_WrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	data := []byte("Secret message")
	encrypted, err := EncryptBytes(data, key1)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = DecryptBytes(encrypted, key2)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key, got nil")
	}
}

func TestDecryptBytes_TooShort(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Data shorter than nonce size
	shortData := []byte("ab")

	_, err = DecryptBytes(shortData, key)
	if err == nil {
		t.Error("Expected error for data shorter than nonce, got nil")
	}
}

func TestEncryptStream_DifferentIV(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	originalData := []byte("Test message")

	// Encrypt twice with same key
	plaintext1 := bytes.NewReader(originalData)
	encrypted1, err := EncryptStream(plaintext1, key)
	if err != nil {
		t.Fatalf("Failed first encryption: %v", err)
	}
	encryptedData1, _ := io.ReadAll(encrypted1)

	plaintext2 := bytes.NewReader(originalData)
	encrypted2, err := EncryptStream(plaintext2, key)
	if err != nil {
		t.Fatalf("Failed second encryption: %v", err)
	}
	encryptedData2, _ := io.ReadAll(encrypted2)

	// Encrypted data should be different due to different IVs
	if bytes.Equal(encryptedData1, encryptedData2) {
		t.Error("Expected different encrypted data due to different IVs")
	}
}

func TestValidKeyLengths(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
		valid     bool
	}{
		{"AES-128", 16, true},
		{"AES-192", 24, true},
		{"AES-256", 32, true},
		{"Invalid", 20, false},
	}

	data := []byte("test data")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLength)

			_, err := EncryptBytes(data, key)
			if tt.valid && err != nil {
				t.Errorf("Expected valid key length %d to work, got error: %v", tt.keyLength, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid key length %d to fail, got no error", tt.keyLength)
			}
		})
	}
}
