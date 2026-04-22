// Penjelasan file:
// Lokasi: internal/models/distribusi.go
// Bagian: model
// File: distribusi
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Distribusi merepresentasikan proses pengiriman pangan dari asal ke tujuan.
type Distribusi struct {
	ID              uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	DariKecamatanID uuid.UUID  `gorm:"type:char(36);not null;index" json:"dari_kecamatan_id"`
	KeKecamatanID   uuid.UUID  `gorm:"type:char(36);not null;index" json:"ke_kecamatan_id"`
	KomoditasID     uuid.UUID  `gorm:"type:char(36);not null;index" json:"komoditas_id"`
	JumlahKg        float64    `gorm:"not null" json:"jumlah_kg"`
	Status          string     `gorm:"size:20;default:'terjadwal'" json:"status"` // terjadwal|proses|selesai
	NamaDriver      string     `gorm:"size:100" json:"nama_driver"`
	NamaKendaraan   string     `gorm:"size:100" json:"nama_kendaraan"`
	JadwalBerangkat time.Time  `json:"jadwal_berangkat"`
	ETA             *time.Time `json:"eta,omitempty"`
	CreatedBy       uuid.UUID  `gorm:"type:char(36)" json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DariKecamatan   Kecamatan  `gorm:"foreignKey:DariKecamatanID" json:"dari_kecamatan,omitempty"`
	KeKecamatan     Kecamatan  `gorm:"foreignKey:KeKecamatanID" json:"ke_kecamatan,omitempty"`
	Komoditas       Komoditas  `gorm:"foreignKey:KomoditasID" json:"komoditas,omitempty"`
}

// BeforeCreate mengisi ID otomatis sebelum data distribusi pertama kali disimpan.
func (d *Distribusi) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
