// Penjelasan file:
// Lokasi: internal/models/user.go
// Bagian: model
// File: user
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User merepresentasikan akun yang dapat masuk dan memakai fitur backend.
type User struct {
	ID          uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	Email       string     `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password    string     `gorm:"not null" json:"-"`
	Phone       string     `gorm:"size:20" json:"phone"`
	Role        string     `gorm:"default:'petani'" json:"role"` // admin|petugas|petani
	KecamatanID *uuid.UUID `gorm:"type:char(36)" json:"kecamatan_id,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BeforeCreate membuat ID unik otomatis ketika akun baru disimpan.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
