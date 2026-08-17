package models

import (
	"time"
)

type Chair struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID  string    `gorm:"type:uuid;not null" json:"clinic_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"` // e.g. "Ünit 1 (Cerrahi)"
	Status    string    `gorm:"type:varchar(50);default:'active'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Appointments []Appointment `gorm:"foreignKey:ChairID" json:"appointments"`
}
