package models

import (
	"time"
)

// DentalRecord Represents a single action on the Odontogram (Diş Şeması)
type DentalRecord struct {
	ID          string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	PatientID   string    `gorm:"type:uuid;not null" json:"patient_id"`
	DoctorID    string    `gorm:"type:uuid;not null" json:"doctor_id"`
	TreatmentID string    `gorm:"type:uuid;not null" json:"treatment_id"`
	ToothNumber string    `gorm:"type:varchar(10);not null" json:"tooth_number"` // e.g. "11", "48", "UL"
	Status      string    `gorm:"type:varchar(50);default:'completed'" json:"status"` // planned, completed
	Notes       string    `gorm:"type:text" json:"notes"`
	Date        time.Time `gorm:"type:date;not null" json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Doctor    Doctor    `gorm:"foreignKey:DoctorID" json:"doctor"`
	Treatment Treatment `gorm:"foreignKey:TreatmentID" json:"treatment"`
}
