package models

import (
	"time"
)

type Patient struct {
	ID             string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID       string    `gorm:"type:uuid;not null" json:"clinic_id"`
	UserID         *string   `gorm:"type:uuid" json:"user_id"` // Nullable if the patient doesn't have an online account
	FirstName      string    `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName       string    `gorm:"type:varchar(100);not null" json:"last_name"`
	Phone          string    `gorm:"type:varchar(50)" json:"phone"`
	Email          string    `gorm:"type:varchar(255)" json:"email"`
	MedicalHistory string    `gorm:"type:text" json:"medical_history"` // JSON or raw text
	Allergies      string    `gorm:"type:text" json:"allergies"`
	Balance        float64   `gorm:"type:decimal(10,2);default:0" json:"balance"` // Positive means debt, negative means credit
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	User          *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Clinic        *Clinic         `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
	Appointments  []Appointment   `gorm:"foreignKey:PatientID" json:"appointments"`
	Reviews       []Review        `gorm:"foreignKey:PatientID" json:"reviews"`
	DentalRecords []DentalRecord  `gorm:"foreignKey:PatientID" json:"dental_records"`
	Payments      []Payment       `gorm:"foreignKey:PatientID" json:"payments"`
}
