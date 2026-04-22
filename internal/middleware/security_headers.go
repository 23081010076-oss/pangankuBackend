// Penjelasan file:
// Lokasi: internal/middleware/security_headers.go
// Bagian: middleware
// File: security_headers
// Fungsi utama: File ini menyisipkan pemeriksaan atau aturan tambahan pada request sebelum masuk handler.
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders menambahkan header keamanan dasar untuk mengurangi risiko serangan umum di web.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HSTS - Force HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Hide server info
		c.Header("X-Powered-By", "")

		c.Next()
	}
}
