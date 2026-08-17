package models

import "time"

type City struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Districts []District `gorm:"foreignKey:CityID" json:"districts,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type District struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CityID    uint      `gorm:"not null" json:"city_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
