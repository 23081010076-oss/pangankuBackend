// Doc:
// Tujuan: Menjalankan migrasi tabel utama backend menggunakan koneksi database aktif.
// Dipakai oleh: Bootstrap aplikasi/server setelah database berhasil terkoneksi.
// Dependensi utama: GORM dan model domain aplikasi.
// Fungsi public/utama: RunMigrations.
// Side effect penting: Membuat atau memperbarui skema tabel pada database MySQL.
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

	log.Println("OK Migrations completed successfully")
}
