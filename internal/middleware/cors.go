// Penjelasan file:
// Lokasi: internal/middleware/cors.go
// Bagian: middleware
// File: cors
// Fungsi utama: File ini menyisipkan pemeriksaan atau aturan tambahan pada request sebelum masuk handler.
package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS middleware untuk mengizinkan request dari frontend
// CORS menambahkan header agar frontend bisa mengakses backend dari origin yang diizinkan.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
