// Penjelasan file:
// Lokasi: internal/middleware/rbac.go
// Bagian: middleware
// File: rbac
// Fungsi utama: File ini menyisipkan pemeriksaan atau aturan tambahan pada request sebelum masuk handler.
package middleware

import (
	"github.com/gin-gonic/gin"
)

// RequireRole memastikan hanya role tertentu yang boleh masuk ke endpoint terkait.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")

		// Cek apakah role user ada dalam list roles yang diizinkan
		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(403, gin.H{"error": "Akses ditolak"})
			return
		}

		c.Next()
	}
}
