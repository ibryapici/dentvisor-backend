package models

import (
	"time"
)

type Appointment struct {
	ID          string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID    string    `gorm:"type:uuid;not null" json:"clinic_id"`
	PatientID   string    `gorm:"type:uuid;not null" json:"patient_id"`
	DoctorID    string    `gorm:"type:uuid;not null" json:"doctor_id"`
	TreatmentID string    `gorm:"type:uuid;not null" json:"treatment_id"`
	ChairID     string    `gorm:"type:uuid;not null" json:"chair_id"`
	StartTime   time.Time `gorm:"not null" json:"start_time"`
	EndTime     time.Time `gorm:"not null" json:"end_time"`
	Status      string    `gorm:"type:varchar(50);default:'scheduled'" json:"status"` // scheduled, completed, cancelled, no-show
	Notes       string    `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Patient   Patient   `gorm:"foreignKey:PatientID" json:"patient"`
	Doctor    Doctor    `gorm:"foreignKey:DoctorID" json:"doctor"`
	Treatment Treatment `gorm:"foreignKey:TreatmentID" json:"treatment"`
	Chair     Chair     `gorm:"foreignKey:ChairID" json:"chair"`
}
