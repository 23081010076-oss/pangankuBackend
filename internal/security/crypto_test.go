// Penjelasan file:
// Lokasi: internal/security/crypto_test.go
// Bagian: security
// File: crypto_test
// Fungsi utama: File ini berisi helper keamanan seperti JWT, hash, validasi, atau enkripsi.
package security_test

import (
	"os"
	"testing"

	"github.com/panganku/backend/internal/security"
)

func TestEncryptDecrypt(t *testing.T) {
	// Set dummy AES key 32 byte (64 hex chars)
	os.Setenv("AES_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	
	original := "Kekurangan beras di Kecamatan Babat"
	
	encrypted, err := security.EncryptAES256(original)
	if err != nil {
		t.Fatal("Encrypt error:", err)
	}
	
	if encrypted == original {
		t.Error("Teks tidak terenkripsi")
	}
	
	decrypted, err := security.DecryptAES256(encrypted)
	if err != nil {
		t.Fatal("Decrypt error:", err)
	}
	
	if decrypted != original {
		t.Errorf("Hasil decrypt salah: %s", decrypted)
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	os.Setenv("AES_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	
	encrypted, err := security.EncryptAES256("")
	if err != nil {
		t.Fatal("Encrypt error:", err)
	}
	
	decrypted, err := security.DecryptAES256(encrypted)
	if err != nil {
		t.Fatal("Decrypt error:", err)
	}
	
	if decrypted != "" {
		t.Errorf("Expected empty string, got: %s", decrypted)
	}
}

func TestValidateEmail(t *testing.T) {
	validEmails := []string{
		"user@example.com",
		"test.user@domain.co.id",
		"admin+tag@site.org",
	}
	
	for _, email := range validEmails {
		if !security.ValidateEmail(email) {
			t.Errorf("Email valid ditolak: %s", email)
		}
	}
	
	invalidEmails := []string{
		"bukan-email",
		"@domain.com",
		"user@",
		"",
		"user @domain.com",
	}
	
	for _, email := range invalidEmails {
		if security.ValidateEmail(email) {
			t.Errorf("Email invalid diterima: %s", email)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	// Valid passwords
	validPasswords := []string{
		"Ab1cdefg",
		"Password123",
		"MyP@ssw0rd",
	}
	
	for _, pwd := range validPasswords {
		if err := security.ValidatePassword(pwd); err != nil {
			t.Errorf("Password valid ditolak: %s, error: %v", pwd, err)
		}
	}
	
	// Invalid passwords
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"short", true},               // terlalu pendek
		{"tanpahurufbesar1", true},    // tanpa huruf besar
		{"TanpaAngka", true},          // tanpa angka
		{"", true},                    // kosong
	}
	
	for _, tt := range tests {
		err := security.ValidatePassword(tt.password)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidatePassword(%s) error = %v, wantErr %v", tt.password, err, tt.wantErr)
		}
	}
}

func TestValidateUUID(t *testing.T) {
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}
	
	for _, uuid := range validUUIDs {
		if !security.ValidateUUID(uuid) {
			t.Errorf("UUID valid ditolak: %s", uuid)
		}
	}
	
	invalidUUIDs := []string{
		"bukan-uuid",
		"123",
		"",
		"550e8400-e29b-41d4-a716",
	}
	
	for _, uuid := range invalidUUIDs {
		if security.ValidateUUID(uuid) {
			t.Errorf("UUID invalid diterima: %s", uuid)
		}
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  spaced  ", "spaced"},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"normal text", "normal text"},
		{"", ""},
	}
	
	for _, tt := range tests {
		got := security.SanitizeString(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

