// Doc:
// Tujuan: Mengatur koneksi database GORM dan konfigurasi connection pool backend.
// Dipakai oleh: Bootstrap aplikasi/server saat inisialisasi dependency backend.
// Dependensi utama: GORM MySQL driver, models, konfigurasi environment, dan logger.
// Fungsi public/utama: ConnectDatabase, GetDB, CloseDatabase.
// Side effect penting: Membuka koneksi MySQL, auto-migrate model, dan mengatur pool koneksi.
package config

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB membaca konfigurasi environment lalu membuka koneksi utama ke MySQL.
func ConnectDB() *gorm.DB {
	db, err := gorm.Open(gormmysql.Open(mysqlDSNFromEnv()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Gagal mendapatkan database instance: %v", err)
	}

	// Set connection pool
	sqlDB.SetMaxIdleConns(intEnv("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetMaxOpenConns(intEnv("DB_MAX_OPEN_CONNS", 100))
	sqlDB.SetConnMaxLifetime(time.Duration(intEnv("DB_CONN_MAX_LIFETIME_MINUTES", 60)) * time.Minute)

	log.Println("OK Database terkoneksi")
	return db
}

func mysqlDSNFromEnv() string {
	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	user := strings.TrimSpace(os.Getenv("DB_USER"))
	password := os.Getenv("DB_PASSWORD")
	dbname := strings.TrimSpace(os.Getenv("DB_NAME"))

	if host == "" || user == "" || dbname == "" {
		log.Fatal("DB_HOST, DB_USER, dan DB_NAME wajib diset")
	}
	if port == "" {
		port = "3306"
	}

	locName := strings.TrimSpace(os.Getenv("DB_LOC"))
	if locName == "" {
		locName = "Local"
	}
	loc, err := time.LoadLocation(locName)
	if err != nil {
		log.Fatalf("DB_LOC tidak valid: %v", err)
	}

	cfg := drivermysql.NewConfig()
	cfg.User = user
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(host, port)
	cfg.DBName = dbname
	cfg.ParseTime = true
	cfg.Loc = loc
	cfg.Params = map[string]string{
		"charset": "utf8mb4",
	}

	if tlsConfig := strings.TrimSpace(os.Getenv("DB_TLS")); tlsConfig != "" {
		cfg.TLSConfig = tlsConfig
	}

	return cfg.FormatDSN()
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		log.Fatalf("%s harus berupa angka positif", key)
	}
	return value
}
