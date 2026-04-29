// Doc:
// Tujuan: Mendefinisikan model master komoditas pangan yang dipakai fitur harga, stok, distribusi, dan admin.
// Dipakai oleh: Handler komoditas, analytics, seeder, serta relasi model harga/stok/distribusi/luas lahan.
// Dependensi utama: GORM, UUID, model relasi backend.
// Fungsi public/utama: Komoditas struct, BeforeCreate.
// Side effect penting: AutoMigrate membaca field model ini; perubahan field memengaruhi skema tabel komoditas.
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
	GambarURL string    `gorm:"size:255" json:"gambar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate mengisi ID unik agar data komoditas baru siap dipakai relasi lain.
func (k *Komoditas) BeforeCreate(_ *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}
