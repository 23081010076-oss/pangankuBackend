package config

import (
	"log"

	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	log.Println("Running database migrations...")
	
	err := db.AutoMigrate(
		&models.User{},
		&models.Komoditas{},
		&models.Kecamatan{},
		&models.HargaPasar{},
		&models.StokPangan{},
		&models.LaporanDarurat{},
		&models.AuditLog{},
		&models.Notifikasi{},
		&models.Distribusi{},
	)
	
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	
	log.Println("✓ Migrations completed successfully")
}
