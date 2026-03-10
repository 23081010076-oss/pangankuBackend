package models

import (
	"time"

	"github.com/google/uuid"
)

type StokPangan struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	KomoditasID uuid.UUID `gorm:"type:uuid;not null;index" json:"komoditas_id"`
	KecamatanID uuid.UUID `gorm:"type:uuid;not null;index" json:"kecamatan_id"`
	StokKg      float64   `gorm:"not null" json:"stok_kg"`
	KapasitasKg float64   `gorm:"not null" json:"kapasitas_kg"`
	PetugasID   uuid.UUID `gorm:"type:uuid" json:"petugas_id,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	Komoditas   Komoditas `gorm:"foreignKey:KomoditasID" json:"komoditas,omitempty"`
	Kecamatan   Kecamatan `gorm:"foreignKey:KecamatanID" json:"kecamatan,omitempty"`
}
