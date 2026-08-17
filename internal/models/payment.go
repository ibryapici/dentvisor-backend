package models

import (
	"time"
)

type Payment struct {
	ID            string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID      string    `gorm:"type:uuid;not null" json:"clinic_id"`
	PatientID     string    `gorm:"type:uuid;not null" json:"patient_id"`
	AppointmentID *string   `gorm:"type:uuid" json:"appointment_id"` // Nullable for direct account payments
	DoctorID      *string   `gorm:"type:uuid" json:"doctor_id"`      // To calculate doctor performance/commission
	Amount        float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	PaymentMethod string    `gorm:"type:varchar(50);not null" json:"payment_method"` // cash, credit_card, transfer
	Type          string    `gorm:"type:varchar(50);not null" json:"type"`           // income, refund
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
