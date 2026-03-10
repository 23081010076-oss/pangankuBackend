package models

import (
	"time"

	"github.com/google/uuid"
)

type LaporanDarurat struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PelaporID    uuid.UUID  `gorm:"type:uuid;not null" json:"pelapor_id"`
	KecamatanID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"kecamatan_id"`
	JenisMasalah string     `gorm:"size:100" json:"jenis_masalah"`
	Deskripsi    string     `json:"deskripsi"` // disimpan terenkripsi AES-256
	FotoURL      string     `json:"foto_url,omitempty"`
	Status       string     `gorm:"default:'baru'" json:"status"` // baru|proses|selesai
	Prioritas    int        `gorm:"default:3" json:"prioritas"`
	CreatedAt    time.Time  `json:"created_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}
