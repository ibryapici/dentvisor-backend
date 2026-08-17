package models

import (
	"time"
)

type Treatment struct {
	ID              string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID        string    `gorm:"type:uuid;not null" json:"clinic_id"`
	Name            string    `gorm:"type:varchar(255);not null" json:"name"`
	DefaultDuration int       `gorm:"type:int;not null;default:30" json:"default_duration"` // in minutes
	DefaultPrice    float64   `gorm:"type:decimal(10,2);not null;default:0" json:"default_price"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	Doctors []Doctor `gorm:"many2many:doctor_treatments;" json:"doctors"`
}
