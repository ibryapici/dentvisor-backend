package models

import (
	"time"
)

type Article struct {
	ID         string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID   string    `gorm:"type:uuid;not null" json:"clinic_id"`
	Title      string    `gorm:"type:varchar(255);not null" json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	ImageURL   string    `gorm:"type:varchar(255)" json:"image_url"`
	Published  bool      `gorm:"type:boolean;default:false" json:"published"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
