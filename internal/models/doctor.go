package models

import (
	"time"
)

type Doctor struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	ClinicID  string    `gorm:"type:uuid;not null" json:"clinic_id"`
	Specialty      string    `gorm:"type:varchar(100)" json:"specialty"` // e.g. Ortodontist, Cerrah
	Title          string    `gorm:"type:varchar(50)" json:"title"`     // e.g. Dt., Uzm. Dt.
	CommissionRate float64   `gorm:"type:decimal(5,2);default:30" json:"commission_rate"` // Prim Oranı (örn: %30, %35)
	BaseSalary     float64   `gorm:"type:decimal(10,2);default:0" json:"base_salary"`     // Sabit Maaş (varsa)
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User       User        `gorm:"foreignKey:UserID" json:"user"`
	Treatments []Treatment `gorm:"many2many:doctor_treatments;" json:"treatments"`
}
