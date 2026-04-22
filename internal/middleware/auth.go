// Penjelasan file:
// Lokasi: internal/middleware/auth.go
// Bagian: middleware
// File: auth
// Fungsi utama: File ini menyisipkan pemeriksaan atau aturan tambahan pada request sebelum masuk handler.
package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/panganku/backend/internal/security"
	"github.com/redis/go-redis/v9"
)

// JWTAuth memeriksa token login, memastikan token valid, lalu menyimpan info user ke context.
func JWTAuth(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token tidak ditemukan"})
			return
		}

		// Format: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Format token tidak valid"})
			return
		}

		token := parts[1]

		// Validasi token
		claims, err := security.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token tidak valid"})
			return
		}

		// Cek blacklist di Redis
		ctx := context.Background()
		exists, _ := redisClient.Exists(ctx, "blacklist:"+token).Result()
		if exists > 0 {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token sudah tidak aktif"})
			return
		}

		// Set user info ke context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// GetUserID mengambil ID user yang sebelumnya disimpan middleware auth.
func GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}

// GetEmail mengambil email user dari context request aktif.
func GetEmail(c *gin.Context) string {
	return c.GetString("email")
}

// GetRole mengambil role user dari context agar handler mudah mengecek hak akses.
func GetRole(c *gin.Context) string {
	return c.GetString("role")
}
