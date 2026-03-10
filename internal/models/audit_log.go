package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid" json:"user_id"`
	Action    string    `gorm:"size:50" json:"action"`
	Resource  string    `gorm:"size:100" json:"resource"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
