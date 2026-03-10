package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/panganku/backend/internal/security"
	"github.com/redis/go-redis/v9"
)

// JWTAuth middleware untuk validasi JWT token
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

// Helper functions untuk mengambil user info dari context
func GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}

func GetEmail(c *gin.Context) string {
	return c.GetString("email")
}

func GetRole(c *gin.Context) string {
	return c.GetString("role")
}
