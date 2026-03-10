package middleware

import (
	"github.com/gin-gonic/gin"
)

// RequireRole middleware untuk validasi role user
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
