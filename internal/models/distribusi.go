package models

import (
	"time"

	"github.com/google/uuid"
)

type Distribusi struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DariKecamatanID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"dari_kecamatan_id"`
	KeKecamatanID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"ke_kecamatan_id"`
	KomoditasID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"komoditas_id"`
	JumlahKg         float64    `gorm:"not null" json:"jumlah_kg"`
	Status           string     `gorm:"size:20;default:'terjadwal'" json:"status"` // terjadwal|proses|selesai
	NamaDriver       string     `gorm:"size:100" json:"nama_driver"`
	NamaKendaraan    string     `gorm:"size:100" json:"nama_kendaraan"`
	JadwalBerangkat  time.Time  `json:"jadwal_berangkat"`
	ETA              *time.Time `json:"eta,omitempty"`
	CreatedBy        uuid.UUID  `gorm:"type:uuid" json:"created_by"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DariKecamatan    Kecamatan  `gorm:"foreignKey:DariKecamatanID" json:"dari_kecamatan,omitempty"`
	KeKecamatan      Kecamatan  `gorm:"foreignKey:KeKecamatanID" json:"ke_kecamatan,omitempty"`
	Komoditas        Komoditas  `gorm:"foreignKey:KomoditasID" json:"komoditas,omitempty"`
}
