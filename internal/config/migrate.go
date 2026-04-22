// Penjelasan file:
// Lokasi: internal/config/migrate.go
// Bagian: config
// File: migrate
// Fungsi utama: File ini mengatur koneksi, konfigurasi, migrasi, atau seed data backend.
package config

import (
	"log"

	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

// AutoMigrate membuat atau menyesuaikan struktur tabel agar sesuai dengan model yang aktif.
func AutoMigrate(db *gorm.DB) {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Komoditas{},
		&models.Kecamatan{},
		&models.HargaPasar{},
		&models.StokPangan{},
		&models.LuasLahan{},
		&models.LaporanDarurat{},
		&models.AuditLog{},
		&models.Notifikasi{},
		&models.Distribusi{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("âœ“ Migrations completed successfully")
}
