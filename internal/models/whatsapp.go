package models

import (
	"time"
)

type WhatsappSetting struct {
	ClinicID             string    `gorm:"type:uuid;primaryKey" json:"clinic_id"`
	PhoneNumber          string    `gorm:"type:varchar(50)" json:"phone_number"`
	Status               string    `gorm:"type:varchar(50);default:'disconnected'" json:"status"` // disconnected, connecting, connected, qr_ready
	QRCode               string    `gorm:"type:text" json:"qr_code"`
	IsAiEnabled          bool      `gorm:"default:true" json:"is_ai_enabled"`
	AiBotName            string    `gorm:"type:varchar(100);default:'Dentvisör Akıllı Klinik Asistanı'" json:"ai_bot_name"`
	AiInstructions       string    `gorm:"type:text" json:"ai_instructions"` // Kliniğe özel bilgi tabanı
	AutoReminderEnabled  bool      `gorm:"default:true" json:"auto_reminder_enabled"`
	ReminderHoursBefore  int       `gorm:"default:24" json:"reminder_hours_before"`
	AiReminderTone       string    `gorm:"type:varchar(50);default:'warm_professional'" json:"ai_reminder_tone"` // warm_professional, formal, friendly
	PostCareEnabled      bool      `gorm:"default:true" json:"post_care_enabled"`
	GoogleReviewLink     string    `gorm:"type:varchar(255)" json:"google_review_link"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// İlişkiler
	Clinic *Clinic `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
}

type WhatsappMessageLog struct {
	ID            string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID      string    `gorm:"type:uuid;not null" json:"clinic_id"`
	PatientID     *string   `gorm:"type:uuid" json:"patient_id"`
	Direction     string    `gorm:"type:varchar(20);not null" json:"direction"` // incoming, outgoing
	Phone         string    `gorm:"type:varchar(50);not null" json:"phone"`
	SenderName    string    `gorm:"type:varchar(100)" json:"sender_name"`
	Message       string    `gorm:"type:text;not null" json:"message"`
	IsAiGenerated bool      `gorm:"default:false" json:"is_ai_generated"`
	Status        string    `gorm:"type:varchar(50);default:'sent'" json:"status"` // sent, delivered, read, received
	CreatedAt     time.Time `json:"created_at"`

	// İlişkiler
	Patient *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}
