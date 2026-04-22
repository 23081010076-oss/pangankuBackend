// Penjelasan file:
// Lokasi: internal/config/redis.go
// Bagian: config
// File: redis
// Fungsi utama: File ini mengatur koneksi, konfigurasi, migrasi, atau seed data backend.
package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis membuka koneksi ke Redis untuk cache, token blacklist, dan kebutuhan cepat lainnya.
func ConnectRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Gagal parse REDIS_URL: %v", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Gagal koneksi ke Redis: %v", err)
	}

	log.Println("âœ“ Redis terkoneksi")
	return client
}
