package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
)

// GenerateKey generates a random 256-bit key
func GenerateKey() ([]byte, error) {
	// 1. Create a 32-byte (256-bit) slice
	key := make([]byte, 32)

	// 2. Use crypto/rand.Read() to fill it with random bytes
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	// 3. Return the key
	return key, nil
}

// EncryptStream creates a streaming encryptor for large files
func EncryptStream(plaintext io.Reader, key []byte) (io.Reader, error) {
	// 1. Create AES cipher block using aes.NewCipher(key)
	// 2. Create GCM mode using cipher.NewGCM(block) or similar stream cipher
	//    *Note:* Standard GCM is authenticated but not stream-friendly for HUGE files in one go without loading into RAM.
	//    For true streaming of huge files with minimal RAM, consider:
	//    - Using AES-CTR (Counter Mode) which turns block cipher into stream cipher.
	//    - OR chunking the file and encrypting each chunk with GCM (more complex but authenticated).
	//
	//    *Simpler Approach for this project (Stream Cipher):*
	//    1. block, _ := aes.NewCipher(key)
	//    2. iv := make([]byte, aes.BlockSize)
	//    3. io.ReadFull(rand.Reader, iv)
	//    4. stream := cipher.NewCTR(block, iv)
	//    5. Return a reader that:
	//       - Reads chunk from plaintext
	//       - XORs with stream
	//       - Prepend IV to output stream
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatalln(err)
		return nil, err
	}
	stream := cipher.NewCTR(block, iv)

	// Create a pipe to connect plaintext reader and encrypted output
	pr, pw := io.Pipe()

	// Write IV first to the pipe
	go func() {
		defer pw.Close()
		if _, err := pw.Write(iv); err != nil {
			log.Fatalln(err)
			return
		}
		buf := make([]byte, 4096)
		for {
			n, err := plaintext.Read(buf)
			if n > 0 {
				encrypted := make([]byte, n)
				stream.XORKeyStream(encrypted, buf[:n])
				if _, err := pw.Write(encrypted); err != nil {
					log.Fatalln(err)
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalln(err)
				return
			}
		}
	}()

	return pr, nil
}
