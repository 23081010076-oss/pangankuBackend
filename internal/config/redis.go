package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

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

	log.Println("✓ Redis terkoneksi")
	return client
}
