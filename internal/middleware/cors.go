// Penjelasan file:
// Lokasi: internal/middleware/cors.go
// Bagian: middleware
// File: cors
// Fungsi utama: File ini menyisipkan pemeriksaan atau aturan tambahan pada request sebelum masuk handler.
package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS middleware untuk mengizinkan request dari frontend
// CORS menambahkan header agar frontend bisa mengakses backend dari origin yang diizinkan.
func CORS() gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins()

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			if origin != "" && !isOriginAllowed(origin, allowedOrigins) {
				c.AbortWithStatusJSON(403, gin.H{"error": "Origin tidak diizinkan"})
				return
			}
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" && os.Getenv("APP_ENV") != "production" {
		raw = "http://localhost:3000,http://localhost:5173,http://localhost:8080,http://127.0.0.1:3000,http://127.0.0.1:5173,http://127.0.0.1:8080"
	}

	allowed := make(map[string]bool)
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

func isOriginAllowed(origin string, allowedOrigins map[string]bool) bool {
	return allowedOrigins[origin]
}
