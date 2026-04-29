// Penjelasan file:
// Lokasi: internal/config/redis.go
// Bagian: config
// File: redis
// Fungsi utama: File ini mengatur koneksi, konfigurasi, migrasi, atau seed data backend.
package config

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis membuka koneksi ke Redis untuk cache, token blacklist, dan kebutuhan cepat lainnya.
func ConnectRedis() *redis.Client {
	client := redis.NewClient(redisOptionsFromEnv())

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Gagal koneksi ke Redis: %v", err)
	}

	log.Println("Redis terkoneksi")
	return client
}

func redisOptionsFromEnv() *redis.Options {
	if redisURL := strings.TrimSpace(os.Getenv("REDIS_URL")); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Gagal parse REDIS_URL: %v", err)
		}
		return opt
	}

	host := strings.TrimSpace(os.Getenv("REDIS_HOST"))
	if host == "" {
		host = "localhost"
	}

	port := strings.TrimSpace(os.Getenv("REDIS_PORT"))
	if port == "" {
		port = "6379"
	}

	db := 0
	if rawDB := strings.TrimSpace(os.Getenv("REDIS_DB")); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			log.Fatalf("REDIS_DB tidak valid: %v", err)
		}
		db = parsedDB
	}

	return &redis.Options{
		Addr:     net.JoinHostPort(host, port),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
}
