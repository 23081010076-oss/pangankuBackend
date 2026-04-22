// Penjelasan file:
// Lokasi: internal/models/luas_lahan.go
// Bagian: model
// File: luas_lahan
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LuasLahan menyimpan data luas tanam per komoditas, kecamatan, dan tahun.
type LuasLahan struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	KomoditasID uuid.UUID `gorm:"type:char(36);not null;index:idx_luas_lahan_unique,unique" json:"komoditas_id"`
	KecamatanID uuid.UUID `gorm:"type:char(36);not null;index:idx_luas_lahan_unique,unique" json:"kecamatan_id"`
	LuasHa      float64   `gorm:"not null" json:"luas_ha"`
	Tahun       int       `gorm:"not null;default:2026;index:idx_luas_lahan_unique,unique" json:"tahun"`
	PetugasID   uuid.UUID `gorm:"type:char(36)" json:"petugas_id,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	Komoditas   Komoditas `gorm:"foreignKey:KomoditasID" json:"komoditas,omitempty"`
	Kecamatan   Kecamatan `gorm:"foreignKey:KecamatanID" json:"kecamatan,omitempty"`
}

// BeforeCreate memastikan data luas lahan memiliki ID unik saat pertama kali dibuat.
func (l *LuasLahan) BeforeCreate(_ *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
