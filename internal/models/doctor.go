package models

import (
	"time"
)

type Doctor struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	ClinicID  string    `gorm:"type:uuid;not null" json:"clinic_id"`
	Specialty string    `gorm:"type:varchar(100)" json:"specialty"` // e.g. Ortodontist, Cerrah
	Title     string    `gorm:"type:varchar(50)" json:"title"`     // e.g. Dt., Uzm. Dt.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User       User        `gorm:"foreignKey:UserID" json:"user"`
	Treatments []Treatment `gorm:"many2many:doctor_treatments;" json:"treatments"`
}
