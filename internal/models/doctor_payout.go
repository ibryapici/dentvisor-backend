package models

import (
	"time"
)

type DoctorPayout struct {
	ID             string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClinicID       string    `gorm:"type:uuid;not null" json:"clinic_id"`
	DoctorID       string    `gorm:"type:uuid;not null" json:"doctor_id"`
	PeriodMonth    int       `gorm:"not null" json:"period_month"` // 1-12
	PeriodYear     int       `gorm:"not null" json:"period_year"`  // Örn: 2026
	TotalRevenue   float64   `gorm:"type:decimal(10,2);default:0" json:"total_revenue"` // Hekimin getirdiği ciro
	CommissionRate float64   `gorm:"type:decimal(5,2);default:0" json:"commission_rate"` // Uygulanan prim %
	CommissionEarn float64   `gorm:"type:decimal(10,2);default:0" json:"commission_earn"` // Hak edilen prim tutarı
	BaseSalary     float64   `gorm:"type:decimal(10,2);default:0" json:"base_salary"`
	BonusAmount    float64   `gorm:"type:decimal(10,2);default:0" json:"bonus_amount"`
	Deductions     float64   `gorm:"type:decimal(10,2);default:0" json:"deductions"` // Kesintiler (lab / avans vb.)
	NetPayout      float64   `gorm:"type:decimal(10,2);default:0" json:"net_payout"` // Toplam ödenecek net
	Status         string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, paid, cancelled
	PaidAt         *time.Time `json:"paid_at"`
	PaymentMethod  string    `gorm:"type:varchar(50)" json:"payment_method"` // bank_transfer, cash
	Notes          string    `gorm:"type:text" json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// İlişkiler
	Doctor *Doctor `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
}
