// Penjelasan file:
// Lokasi: internal/models/komoditas.go
// Bagian: model
// File: komoditas
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Komoditas menyimpan data master bahan pangan yang dipantau aplikasi.
type Komoditas struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Nama      string    `gorm:"size:100;not null" json:"nama"`
	Satuan    string    `gorm:"default:'kg'" json:"satuan"`
	Kategori  string    `gorm:"size:50" json:"kategori"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate mengisi ID unik agar data komoditas baru siap dipakai relasi lain.
func (k *Komoditas) BeforeCreate(_ *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}
