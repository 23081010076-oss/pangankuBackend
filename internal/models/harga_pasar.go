package models

import (
	"time"

	"github.com/google/uuid"
)

type HargaPasar struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	KomoditasID uuid.UUID `gorm:"type:uuid;not null;index" json:"komoditas_id"`
	KecamatanID uuid.UUID `gorm:"type:uuid;not null;index" json:"kecamatan_id"`
	HargaPerKg  float64   `gorm:"not null" json:"harga_per_kg"`
	Tanggal     time.Time `gorm:"index" json:"tanggal"`
	CreatedBy   uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Komoditas   Komoditas `gorm:"foreignKey:KomoditasID" json:"komoditas,omitempty"`
	Kecamatan   Kecamatan `gorm:"foreignKey:KecamatanID" json:"kecamatan,omitempty"`
}
