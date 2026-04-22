// Penjelasan file:
// Lokasi: internal/models/harga_pasar.go
// Bagian: model
// File: harga_pasar
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HargaPasar menyimpan catatan harga komoditas pada wilayah dan tanggal tertentu.
type HargaPasar struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	KomoditasID uuid.UUID `gorm:"type:char(36);not null;index" json:"komoditas_id"`
	KecamatanID uuid.UUID `gorm:"type:char(36);not null;index" json:"kecamatan_id"`
	HargaPerKg  float64   `gorm:"not null" json:"harga_per_kg"`
	Tanggal     time.Time `gorm:"index" json:"tanggal"`
	CreatedBy   uuid.UUID `gorm:"type:char(36)" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Komoditas   Komoditas `gorm:"foreignKey:KomoditasID" json:"komoditas,omitempty"`
	Kecamatan   Kecamatan `gorm:"foreignKey:KecamatanID" json:"kecamatan,omitempty"`
}

// BeforeCreate memastikan setiap catatan harga memiliki ID unik sebelum masuk database.
func (h *HargaPasar) BeforeCreate(_ *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}
