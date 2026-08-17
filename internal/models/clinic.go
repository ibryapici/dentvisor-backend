package models

import (
	"time"
)

type Clinic struct {
	ID           string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Phone        string    `gorm:"type:varchar(50)" json:"phone"`
	Address      string    `gorm:"type:text" json:"address"`
	AboutText    string    `gorm:"type:text" json:"about_text"`
	WorkingHours string    `gorm:"type:varchar(100)" json:"working_hours"`
	LogoURL      string    `gorm:"type:varchar(255)" json:"logo_url"`
	LocationURL  string    `gorm:"type:varchar(255)" json:"location_url"`

	// Lokasyon Bilgileri
	CityID       *uint     `json:"city_id"`
	DistrictID   *uint     `json:"district_id"`
	City         *City     `gorm:"foreignKey:CityID" json:"city,omitempty"`
	District     *District `gorm:"foreignKey:DistrictID" json:"district,omitempty"`

	// İlişkiler
	Users        []User        `gorm:"foreignKey:ClinicID" json:"users,omitempty"`
	Doctors      []Doctor      `gorm:"foreignKey:ClinicID" json:"doctors,omitempty"`
	Patients     []Patient     `gorm:"foreignKey:ClinicID" json:"patients,omitempty"`
	Treatments   []Treatment   `gorm:"foreignKey:ClinicID" json:"treatments,omitempty"`
	Chairs       []Chair       `gorm:"foreignKey:ClinicID" json:"chairs,omitempty"`
	Reviews      []Review      `gorm:"foreignKey:ClinicID" json:"reviews"`
	Articles     []Article     `gorm:"foreignKey:ClinicID" json:"articles"`
	Appointments []Appointment `gorm:"foreignKey:ClinicID" json:"appointments,omitempty"`

	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
