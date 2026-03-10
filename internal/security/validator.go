package security

import (
	"errors"
	"html"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// ValidateEmail memvalidasi format email
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidatePassword memvalidasi kekuatan password
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("Password minimal 8 karakter")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper {
		return errors.New("Password harus mengandung huruf kapital")
	}
	if !hasDigit {
		return errors.New("Password harus mengandung angka")
	}

	return nil
}

// ValidateUUID memvalidasi format UUID
func ValidateUUID(id string) bool {
	return uuidRegex.MatchString(strings.ToLower(id))
}

// SanitizeString membersihkan string dari HTML/script injection
func SanitizeString(input string) string {
	// Trim whitespace
	s := strings.TrimSpace(input)
	
	// Escape HTML entities
	s = html.EscapeString(s)
	
	return s
}
