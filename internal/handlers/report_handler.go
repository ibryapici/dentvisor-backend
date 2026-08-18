package handlers

import (
	"net/http"
	"strconv"
	"time"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// GET /api/protected/admin/reports/performance
func (h *ReportHandler) GetPerformanceReport(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	// 1. Doktor performans raporu
	type DoctorPerformance struct {
		DoctorID      string  `json:"doctor_id"`
		DoctorName    string  `json:"doctor_name"`
		TotalPatients int64   `json:"total_patients"`
		TotalRevenue  float64 `json:"total_revenue"`
	}
	var docStats []DoctorPerformance

	database.DB.Raw(`
		SELECT 
			d.id as doctor_id, 
			u.first_name || ' ' || u.last_name as doctor_name,
			COUNT(DISTINCT a.patient_id) as total_patients,
			COALESCE(SUM(t.default_price), 0) as total_revenue
		FROM doctors d
		JOIN users u ON d.user_id = u.id
		LEFT JOIN appointments a ON a.doctor_id = d.id AND a.status = 'completed' AND a.clinic_id = ?
		LEFT JOIN treatments t ON a.treatment_id = t.id
		WHERE d.clinic_id = ?
		GROUP BY d.id, u.first_name, u.last_name
	`, clinicID, clinicID).Scan(&docStats)

	// 2. Tedavi istatistikleri
	type TreatmentStats struct {
		TreatmentName string `json:"name"`
		Count         int64  `json:"count"`
	}
	var trtStats []TreatmentStats

	database.DB.Raw(`
		SELECT 
			t.name as treatment_name,
			COUNT(a.id) as count
		FROM treatments t
		JOIN appointments a ON a.treatment_id = t.id AND a.status = 'completed' AND a.clinic_id = ?
		WHERE t.clinic_id = ?
		GROUP BY t.id, t.name
		ORDER BY count DESC
		LIMIT 5
	`, clinicID, clinicID).Scan(&trtStats)

	// 3. Aylık gelir tablosu (Son 6 ay)
	type MonthlyRevenue struct {
		Month   string  `json:"month"`
		Revenue float64 `json:"revenue"`
	}
	var monthlyStats []MonthlyRevenue

	database.DB.Raw(`
		SELECT 
			TO_CHAR(a.start_time, 'Mon') as month,
			COALESCE(SUM(t.default_price), 0) as revenue
		FROM appointments a
		JOIN treatments t ON a.treatment_id = t.id
		WHERE a.status = 'completed' AND a.clinic_id = ? AND a.start_time >= ?
		GROUP BY TO_CHAR(a.start_time, 'Mon'), DATE_TRUNC('month', a.start_time)
		ORDER BY DATE_TRUNC('month', a.start_time) ASC
	`, clinicID, time.Now().AddDate(0, -5, 0)).Scan(&monthlyStats)

	c.JSON(http.StatusOK, gin.H{
		"doctors":         docStats,
		"top_treatments":  trtStats,
		"monthly_revenue": monthlyStats,
	})
}

// GET /api/protected/reports/doctor-commissions (Hekim Prim ve Hakediş Raporu)
func (h *ReportHandler) GetDoctorCommissions(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	// Hekimleri al
	var doctors []models.Doctor
	database.DB.Preload("User").Where("clinic_id = ?", clinicID).Find(&doctors)

	type DoctorCommissionSummary struct {
		DoctorID       string  `json:"doctor_id"`
		DoctorName     string  `json:"doctor_name"`
		Title          string  `json:"title"`
		CommissionRate float64 `json:"commission_rate"`
		BaseSalary     float64 `json:"base_salary"`
		TotalTreatments int64  `json:"total_treatments"`
		TotalRevenue   float64 `json:"total_revenue"`
		CommissionEarn float64 `json:"commission_earn"`
		NetPayout      float64 `json:"net_payout"`
		IsPaid         bool    `json:"is_paid"`
	}

	var summaries []DoctorCommissionSummary

	for _, doc := range doctors {
		var totalRev float64
		var count int64

		// Bu ay tamamlanan randevuların/tedavilerin toplamı
		database.DB.Raw(`
			SELECT 
				COUNT(a.id),
				COALESCE(SUM(t.default_price), 0)
			FROM appointments a
			JOIN treatments t ON a.treatment_id = t.id
			WHERE a.doctor_id = ? 
			  AND a.status = 'completed'
			  AND a.clinic_id = ?
			  AND EXTRACT(MONTH FROM a.start_time) = ?
			  AND EXTRACT(YEAR FROM a.start_time) = ?
		`, doc.ID, clinicID, month, year).Row().Scan(&count, &totalRev)

		rate := doc.CommissionRate
		if rate <= 0 {
			rate = 30.0 // Varsayılan %30
		}

		commissionEarn := (totalRev * rate) / 100
		netPayout := doc.BaseSalary + commissionEarn

		// Bu dönem ödenmiş mi?
		var payoutCount int64
		database.DB.Model(&models.DoctorPayout{}).
			Where("doctor_id = ? AND period_month = ? AND period_year = ? AND status = 'paid'", doc.ID, month, year).
			Count(&payoutCount)

		summaries = append(summaries, DoctorCommissionSummary{
			DoctorID:        doc.ID,
			DoctorName:      doc.User.FirstName + " " + doc.User.LastName,
			Title:           doc.Title,
			CommissionRate:  rate,
			BaseSalary:      doc.BaseSalary,
			TotalTreatments: count,
			TotalRevenue:    totalRev,
			CommissionEarn:  commissionEarn,
			NetPayout:       netPayout,
			IsPaid:          payoutCount > 0,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"period_month": month,
		"period_year":  year,
		"doctors":      summaries,
	})
}

// PUT /api/protected/reports/doctors/:id/commission (Hekim Prim ve Maaş Güncelleme)
func (h *ReportHandler) UpdateDoctorCommission(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	docID := c.Param("id")
	var req struct {
		CommissionRate float64 `json:"commission_rate"`
		BaseSalary     float64 `json:"base_salary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	if err := database.DB.Model(&models.Doctor{}).
		Where("id = ? AND clinic_id = ?", docID, clinicID).
		Updates(map[string]interface{}{
			"commission_rate": req.CommissionRate,
			"base_salary":     req.BaseSalary,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hekim prim ve maaş ayarları güncellendi"})
}

// POST /api/protected/reports/doctor-payouts (Hakediş Ödemesi Kaydet)
func (h *ReportHandler) CreateDoctorPayout(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req struct {
		DoctorID       string  `json:"doctor_id" binding:"required"`
		PeriodMonth    int     `json:"period_month" binding:"required"`
		PeriodYear     int     `json:"period_year" binding:"required"`
		TotalRevenue   float64 `json:"total_revenue"`
		CommissionRate float64 `json:"commission_rate"`
		CommissionEarn float64 `json:"commission_earn"`
		BaseSalary     float64 `json:"base_salary"`
		BonusAmount    float64 `json:"bonus_amount"`
		Deductions     float64 `json:"deductions"`
		NetPayout      float64 `json:"net_payout"`
		PaymentMethod  string  `json:"payment_method"`
		Notes          string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ödeme bilgileri"})
		return
	}

	now := time.Now()
	payout := models.DoctorPayout{
		ClinicID:       clinicID,
		DoctorID:       req.DoctorID,
		PeriodMonth:    req.PeriodMonth,
		PeriodYear:     req.PeriodYear,
		TotalRevenue:   req.TotalRevenue,
		CommissionRate: req.CommissionRate,
		CommissionEarn: req.CommissionEarn,
		BaseSalary:     req.BaseSalary,
		BonusAmount:    req.BonusAmount,
		Deductions:     req.Deductions,
		NetPayout:      req.NetPayout,
		Status:         "paid",
		PaidAt:         &now,
		PaymentMethod:  req.PaymentMethod,
		Notes:          req.Notes,
	}

	if err := database.DB.Create(&payout).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ödeme kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, payout)
}

// GET /api/protected/reports/doctor-payouts
func (h *ReportHandler) GetDoctorPayouts(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var payouts []models.DoctorPayout
	database.DB.Preload("Doctor.User").
		Where("clinic_id = ?", clinicID).
		Order("created_at desc").
		Find(&payouts)

	c.JSON(http.StatusOK, payouts)
}
