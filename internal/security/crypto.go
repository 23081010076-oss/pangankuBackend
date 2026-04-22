// Penjelasan file:
// Lokasi: internal/security/crypto.go
// Bagian: security
// File: crypto
// Fungsi utama: File ini berisi helper keamanan seperti JWT, hash, validasi, atau enkripsi.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
)

// EncryptAES256 mengenkripsi plaintext dengan AES-256-GCM
func EncryptAES256(plaintext string) (string, error) {
	// Decode hex key dari environment
	keyHex := os.Getenv("AES_KEY")
	if keyHex == "" {
		return "", errors.New("AES_KEY tidak diset")
	}
	
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", errors.New("AES_KEY format tidak valid")
	}
	
	if len(key) != 32 {
		return "", errors.New("AES_KEY harus 64 karakter hex (32 byte)")
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES256 mendekripsi ciphertext yang dienkripsi dengan AES-256-GCM
func DecryptAES256(encoded string) (string, error) {
	// Decode hex key dari environment
	keyHex := os.Getenv("AES_KEY")
	if keyHex == "" {
		return "", errors.New("AES_KEY tidak diset")
	}
	
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", errors.New("AES_KEY format tidak valid")
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext terlalu pendek")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

