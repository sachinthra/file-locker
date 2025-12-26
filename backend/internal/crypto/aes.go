package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// GenerateKey generates a random 256-bit key
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// EncryptStream creates a streaming encryptor for large files
func EncryptStream(plaintext io.Reader, key []byte) (io.Reader, error) {
	// Validate key length before creating cipher
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid AES key length: got %d bytes, need 16, 24, or 32", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		// Write IV first
		if _, err := pw.Write(iv); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to write IV: %w", err))
			return
		}

		buf := make([]byte, 4096)
		for {
			n, err := plaintext.Read(buf)
			if n > 0 {
				stream.XORKeyStream(buf[:n], buf[:n]) // Reuse buffer
				if _, writeErr := pw.Write(buf[:n]); writeErr != nil {
					pw.CloseWithError(fmt.Errorf("failed to write encrypted data: %w", writeErr))
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				pw.CloseWithError(fmt.Errorf("failed to read plaintext: %w", err))
				return
			}
		}
	}()

	return pr, nil
}

// DecryptStream creates a streaming decryptor
func DecryptStream(ciphertext io.Reader, key []byte) (io.Reader, error) {
	// Validate key length before creating cipher
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid AES key length: got %d bytes, need 16, 24, or 32", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(ciphertext, iv); err != nil {
		return nil, fmt.Errorf("failed to read IV: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		buf := make([]byte, 4096)
		for {
			n, err := ciphertext.Read(buf)
			if n > 0 {
				stream.XORKeyStream(buf[:n], buf[:n]) // Reuse buffer
				if _, writeErr := pw.Write(buf[:n]); writeErr != nil {
					pw.CloseWithError(fmt.Errorf("failed to write decrypted data: %w", writeErr))
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				pw.CloseWithError(fmt.Errorf("failed to read ciphertext: %w", err))
				return
			}
		}
	}()

	return pr, nil
}

// EncryptBytes encrypts small data (for keys, metadata, etc.)
func EncryptBytes(plaintext, key []byte) ([]byte, error) {
	// Validate key length before creating cipher
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid AES key length: got %d bytes, need 16, 24, or 32", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptBytes decrypts small data
func DecryptBytes(ciphertext, key []byte) ([]byte, error) {
	// Validate key length before creating cipher
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid AES key length: got %d bytes, need 16, 24, or 32", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
