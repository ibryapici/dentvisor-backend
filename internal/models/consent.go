package models

import (
	"time"
)

// PatientConsent represents a patient's explicit consent for a clinic to view their cross-clinic medical dental history (e-Nabiz model).
type PatientConsent struct {
	ID             string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	PatientID      string    `gorm:"type:uuid;not null" json:"patient_id"`
	TargetClinicID string    `gorm:"type:uuid;not null" json:"target_clinic_id"`
	Status         string    `gorm:"type:varchar(20);default:'pending'" json:"status"` // 'pending', 'approved', 'revoked'
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Patient      *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	TargetClinic *Clinic  `gorm:"foreignKey:TargetClinicID" json:"target_clinic,omitempty"`
}
