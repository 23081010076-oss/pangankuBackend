// Penjelasan file:
// Lokasi: internal/models/audit_log.go
// Bagian: model
// File: audit_log
// Fungsi utama: File ini mendefinisikan struktur data atau tabel yang dipakai backend.
package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog menyimpan jejak aktivitas penting agar perubahan data bisa ditelusuri.
type AuditLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uuid.UUID `gorm:"type:char(36)" json:"user_id"`
	Action    string    `gorm:"size:50" json:"action"`
	Resource  string    `gorm:"size:100" json:"resource"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
