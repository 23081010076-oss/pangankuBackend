package security_test

import (
	"testing"

	"github.com/panganku/backend/internal/security"
)

func TestHashPassword(t *testing.T) {
	password := "MyPassword123"
	
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatal("HashPassword error:", err)
	}
	
	if hash == password {
		t.Error("Password tidak di-hash")
	}
	
	// Hash harus berisi ":" sebagai separator
	if len(hash) < 10 {
		t.Error("Hash terlalu pendek")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "MyPassword123"
	
	hash, _ := security.HashPassword(password)
	
	// Test password yang benar
	if !security.VerifyPassword(password, hash) {
		t.Error("Password valid tidak terverifikasi")
	}
	
	// Test password yang salah
	if security.VerifyPassword("WrongPassword", hash) {
		t.Error("Password salah terverifikasi")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := security.GenerateRefreshToken()
	if err != nil {
		t.Fatal("GenerateRefreshToken error:", err)
	}
	
	if len(token1) < 20 {
		t.Error("Refresh token terlalu pendek")
	}
	
	// Generate lagi, harus berbeda
	token2, _ := security.GenerateRefreshToken()
	if token1 == token2 {
		t.Error("Refresh token harus unik")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "MyPassword123"
	for i := 0; i < b.N; i++ {
		security.HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "MyPassword123"
	hash, _ := security.HashPassword(password)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		security.VerifyPassword(password, hash)
	}
}
