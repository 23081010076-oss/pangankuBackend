// Penjelasan file:
// Lokasi: internal/models/stok_pangan.go
// Bagian: model
// File: stok_pangan
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StokPangan menyimpan jumlah ketersediaan komoditas pada wilayah dan waktu tertentu.
type StokPangan struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	KomoditasID uuid.UUID `gorm:"type:char(36);not null;index" json:"komoditas_id"`
	KecamatanID uuid.UUID `gorm:"type:char(36);not null;index" json:"kecamatan_id"`
	StokKg      float64   `gorm:"not null" json:"stok_kg"`
	KapasitasKg float64   `gorm:"not null" json:"kapasitas_kg"`
	PetugasID   uuid.UUID `gorm:"type:char(36)" json:"petugas_id,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	Komoditas   Komoditas `gorm:"foreignKey:KomoditasID" json:"komoditas,omitempty"`
	Kecamatan   Kecamatan `gorm:"foreignKey:KecamatanID" json:"kecamatan,omitempty"`
}

// BeforeCreate menyiapkan ID unik untuk catatan stok baru.
func (s *StokPangan) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
