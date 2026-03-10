package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	Email       string     `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password    string     `gorm:"not null" json:"-"`
	Phone       string     `gorm:"size:20" json:"phone"`
	Role        string     `gorm:"default:'publik'" json:"role"` // admin|petugas|petani|publik
	KecamatanID *uuid.UUID `gorm:"type:uuid" json:"kecamatan_id,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
