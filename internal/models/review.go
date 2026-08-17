package models

import (
	"time"
)

type Review struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID  string    `gorm:"type:uuid;not null" json:"clinic_id"`
	PatientID string    `gorm:"type:uuid;not null" json:"patient_id"`
	Rating    int       `gorm:"type:int;not null" json:"rating"` // 1 to 5
	Comment   string    `gorm:"type:text" json:"comment"`
	Status    string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Patient Patient `gorm:"foreignKey:PatientID" json:"patient"`
}
