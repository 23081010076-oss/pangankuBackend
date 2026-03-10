package models

import (
	"time"

	"github.com/google/uuid"
)

type Komoditas struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nama      string    `gorm:"size:100;not null" json:"nama"`
	Satuan    string    `gorm:"default:'kg'" json:"satuan"`
	Kategori  string    `gorm:"size:50" json:"kategori"`
	CreatedAt time.Time `json:"created_at"`
}
