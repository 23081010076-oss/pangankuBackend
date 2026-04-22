// Penjelasan file:
// Lokasi: internal/models/laporan_darurat.go
// Bagian: model
// File: laporan_darurat
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LaporanDarurat menyimpan laporan kejadian lapangan yang butuh perhatian cepat.
type LaporanDarurat struct {
	ID           uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	PelaporID    uuid.UUID  `gorm:"type:char(36);not null" json:"pelapor_id"`
	KecamatanID  uuid.UUID  `gorm:"type:char(36);not null;index" json:"kecamatan_id"`
	JenisMasalah string     `gorm:"size:100" json:"jenis_masalah"`
	Deskripsi    string     `json:"deskripsi"` // disimpan terenkripsi AES-256
	FotoURL      string     `json:"foto_url,omitempty"`
	Status       string     `gorm:"default:'baru'" json:"status"` // baru|proses|selesai
	Prioritas    int        `gorm:"default:3" json:"prioritas"`
	CreatedAt    time.Time  `json:"created_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

// BeforeCreate menyiapkan ID unik sebelum laporan darurat dibuat.
func (l *LaporanDarurat) BeforeCreate(_ *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
