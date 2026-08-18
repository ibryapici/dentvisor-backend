package models

import (
	"time"
)

type Laboratory struct {
	ID            string     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID      string     `gorm:"type:uuid;not null" json:"clinic_id"`
	Name          string     `gorm:"type:varchar(255);not null" json:"name"`
	Phone         string     `gorm:"type:varchar(50)" json:"phone"`
	ContactPerson string     `gorm:"type:varchar(100)" json:"contact_person"`
	Address       string     `gorm:"type:text" json:"address"`
	Notes         string     `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// İlişkiler
	Orders []LabOrder `gorm:"foreignKey:LaboratoryID" json:"orders,omitempty"`
}

type LabOrder struct {
	ID            string     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID      string     `gorm:"type:uuid;not null" json:"clinic_id"`
	PatientID     string     `gorm:"type:uuid;not null" json:"patient_id"`
	DoctorID      *string    `gorm:"type:uuid" json:"doctor_id"`
	LaboratoryID  string     `gorm:"type:uuid;not null" json:"laboratory_id"`
	WorkType      string     `gorm:"type:varchar(100);not null" json:"work_type"` // Zirkonyum, E-Max, Porselen, Gece Plağı, Total Protez vb.
	ToothNumbers  string     `gorm:"type:varchar(100)" json:"tooth_numbers"`     // Örn: "11, 21, 22"
	ShadeColor    string     `gorm:"type:varchar(50)" json:"shade_color"`        // A1, A2, B1, BL2 vb.
	Status        string     `gorm:"type:varchar(50);default:'sent'" json:"status"` // sent, in_lab, try_in, completed, cancelled
	SentDate      time.Time  `json:"sent_date"`
	DueDate       *time.Time `json:"due_date"`
	CompletedDate *time.Time `json:"completed_date"`
	Cost          float64    `gorm:"type:decimal(10,2);default:0" json:"cost"`
	IsPaid        bool       `gorm:"default:false" json:"is_paid"`
	Notes         string     `gorm:"type:text" json:"notes"` // Teknisyen notları
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// İlişkiler
	Patient    *Patient    `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Doctor     *Doctor     `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Laboratory *Laboratory `gorm:"foreignKey:LaboratoryID" json:"laboratory,omitempty"`
}
