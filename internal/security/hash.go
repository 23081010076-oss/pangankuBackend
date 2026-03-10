package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength = 16
	keyLength  = 32
	timeCost   = 1
	memory     = 64 * 1024
	threads    = 4
)

// HashPassword menggunakan argon2id untuk hash password
func HashPassword(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Hash password dengan argon2
	hash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, keyLength)

	// Encode: base64(salt):base64(hash)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	encodedHash := base64.StdEncoding.EncodeToString(hash)
	
	return encodedSalt + ":" + encodedHash, nil
}

// VerifyPassword memverifikasi password dengan hash yang tersimpan
func VerifyPassword(password, encoded string) bool {
	// Split encoded string
	parts := strings.Split(encoded, ":")
	if len(parts) != 2 {
		return false
	}

	// Decode salt dan hash
	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	
	storedHash, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	// Hitung hash dengan salt yang sama
	computedHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, keyLength)

	// Compare hashes
	if len(computedHash) != len(storedHash) {
		return false
	}
	
	for i := range computedHash {
		if computedHash[i] != storedHash[i] {
			return false
		}
	}
	
	return true
}

// GenerateRefreshToken membuat refresh token random
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("gagal generate refresh token")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
