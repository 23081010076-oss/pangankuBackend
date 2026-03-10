package models

import (
	"time"

	"github.com/google/uuid"
)

type Notifikasi struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Judul     string    `gorm:"size:200;not null" json:"judul"`
	Isi       string    `gorm:"not null" json:"isi"`
	Tipe      string    `gorm:"size:50;default:'info'" json:"tipe"` // info|warning|error|success
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	DeepLink  string    `gorm:"size:200" json:"deep_link,omitempty"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
