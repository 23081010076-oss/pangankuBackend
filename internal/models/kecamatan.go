package models

import (
	"github.com/google/uuid"
)

type Kecamatan struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nama   string    `gorm:"size:100;not null" json:"nama"`
	Lat    float64   `json:"lat"`
	Lng    float64   `json:"lng"`
	LuasHa float64   `json:"luas_ha"`
}
