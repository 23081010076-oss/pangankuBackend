// Penjelasan file:
// Lokasi: internal/models/notifikasi.go
// Bagian: model
// File: notifikasi
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notifikasi menyimpan pesan sistem yang akan ditampilkan ke pengguna aplikasi.
type Notifikasi struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	Judul     string    `gorm:"size:200;not null" json:"judul"`
	Isi       string    `gorm:"not null" json:"isi"`
	Tipe      string    `gorm:"size:50;default:'info'" json:"tipe"` // info|warning|error|success
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	DeepLink  string    `gorm:"size:200" json:"deep_link,omitempty"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// BeforeCreate mengisi ID unik untuk notifikasi baru sebelum disimpan.
func (n *Notifikasi) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
