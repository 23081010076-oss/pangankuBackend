package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders middleware untuk menambahkan security headers
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
