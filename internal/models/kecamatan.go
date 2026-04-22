// Penjelasan file:
// Lokasi: internal/models/kecamatan.go
// Bagian: model
// File: kecamatan
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Kecamatan menyimpan data wilayah administratif yang dipakai di banyak modul aplikasi.
type Kecamatan struct {
	ID     uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Nama   string    `gorm:"size:100;not null" json:"nama"`
	Lat    float64   `json:"lat"`
	Lng    float64   `json:"lng"`
	LuasHa float64   `json:"luas_ha"`
}

// BeforeCreate menyiapkan ID unik untuk kecamatan baru sebelum disimpan.
func (k *Kecamatan) BeforeCreate(_ *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}
