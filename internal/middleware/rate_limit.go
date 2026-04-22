// Penjelasan file:
// Lokasi: internal/middleware/rate_limit.go
// Bagian: middleware
// File: rate_limit
// Fungsi utama: File ini menyisipkan pemeriksaan atau aturan tambahan pada request sebelum masuk handler.
package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit middleware untuk membatasi request berdasarkan IP dan path
// RateLimit membatasi jumlah request per klien dalam jendela waktu tertentu untuk mencegah spam.
func RateLimit(rdb *redis.Client, maxReq int, windowSec int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Key berdasarkan IP dan path
		key := fmt.Sprintf("ratelimit:%s:%s", c.ClientIP(), c.FullPath())

		// Increment counter
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next() // Jika Redis error, lanjutkan request
			return
		}

		// Set expire untuk window pertama kali
		if count == 1 {
			rdb.Expire(ctx, key, time.Duration(windowSec)*time.Second)
		}

		// Cek apakah melebihi limit
		if count > int64(maxReq) {
			c.Header("Retry-After", strconv.Itoa(windowSec))
			c.AbortWithStatusJSON(429, gin.H{
				"error":       "Terlalu banyak request",
				"retry_after": windowSec,
			})
			return
		}

		c.Next()
	}
}
