// Doc:
// Tujuan: Mengatur koneksi database GORM dan konfigurasi connection pool backend.
// Dipakai oleh: Bootstrap aplikasi/server saat inisialisasi dependency backend.
// Dependensi utama: GORM MySQL driver, models, konfigurasi environment, dan logger.
// Fungsi public/utama: ConnectDatabase, GetDB, CloseDatabase.
// Side effect penting: Membuka koneksi MySQL, auto-migrate model, dan mengatur pool koneksi.
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB membaca konfigurasi environment lalu membuka koneksi utama ke MySQL.
func ConnectDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
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
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("OK Database terkoneksi")
	return db
}
