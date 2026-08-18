package models

import (
	"time"
)

type TreatmentPlan struct {
	ID             string              `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID       string              `gorm:"type:uuid;not null" json:"clinic_id"`
	PatientID      string              `gorm:"type:uuid;not null" json:"patient_id"`
	DoctorID       *string             `gorm:"type:uuid" json:"doctor_id"`
	Title          string              `gorm:"type:varchar(255);not null" json:"title"` // Örn: "Genel İmplant & Gülüş Tasarımı"
	Status         string              `gorm:"type:varchar(50);default:'draft'" json:"status"` // draft, presented, accepted, rejected, completed
	Subtotal       float64             `gorm:"type:decimal(10,2);default:0" json:"subtotal"`
	DiscountAmount float64             `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	DiscountRate   float64             `gorm:"type:decimal(5,2);default:0" json:"discount_rate"` // Yüzde indirim
	TotalAmount    float64             `gorm:"type:decimal(10,2);default:0" json:"total_amount"`
	Notes          string              `gorm:"type:text" json:"notes"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`

	// İlişkiler
	Patient *Patient            `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Doctor  *Doctor             `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Items   []TreatmentPlanItem `gorm:"foreignKey:TreatmentPlanID;constraint:OnDelete:CASCADE" json:"items"`
}

type TreatmentPlanItem struct {
	ID              string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TreatmentPlanID string    `gorm:"type:uuid;not null" json:"treatment_plan_id"`
	ToothNumber     *int      `json:"tooth_number"` // Örn: 11, 21, 46 (boş ise genel tedavi)
	TreatmentID     *string   `gorm:"type:uuid" json:"treatment_id"`
	TreatmentName   string    `gorm:"type:varchar(255);not null" json:"treatment_name"`
	UnitPrice       float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Quantity        int       `gorm:"default:1" json:"quantity"`
	TotalPrice      float64   `gorm:"type:decimal(10,2);not null" json:"total_price"`
	Notes           string    `gorm:"type:text" json:"notes"`
	Status          string    `gorm:"type:varchar(50);default:'planned'" json:"status"` // planned, in_progress, completed, cancelled
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
