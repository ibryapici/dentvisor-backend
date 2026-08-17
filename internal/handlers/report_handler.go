package handlers

import (
	"net/http"
	"time"

	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// GET /api/protected/reports/performance
func (h *ReportHandler) GetPerformanceReport(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	// 1. Doktor performans raporu (Hangi doktor, kaç hasta bakmış, toplam ciro ne kadar?)
	type DoctorPerformance struct {
		DoctorID      string  `json:"doctor_id"`
		DoctorName    string  `json:"doctor_name"`
		TotalPatients int64   `json:"total_patients"`
		TotalRevenue  float64 `json:"total_revenue"`
	}
	var docStats []DoctorPerformance

	// Sadece tamamlanmış (completed) ödemeler hesaplanır.
	// NOT: Gerçek projede payments tablosunda doctor_id veya appointment ile ilişkili ödemeler hesaplanır.
	// Biz şimdilik randevular tablosundaki tamamlanmış randevulardan tahmini ciro hesaplayalım.
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

	// 2. Tedavi istatistikleri (En çok hangi tedaviler uygulanmış)
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

	// Basit bir yaklaşımla, tamamlanmış randevulardaki tedavilerin fiyatlarını tarihe göre grupluyoruz.
	database.DB.Raw(`
		SELECT 
			TO_CHAR(a.start_time, 'Mon') as month,
			SUM(t.default_price) as revenue
		FROM appointments a
		JOIN treatments t ON a.treatment_id = t.id
		WHERE a.status = 'completed' AND a.clinic_id = ? AND a.start_time >= ?
		GROUP BY TO_CHAR(a.start_time, 'Mon'), DATE_TRUNC('month', a.start_time)
		ORDER BY DATE_TRUNC('month', a.start_time) ASC
	`, clinicID, time.Now().AddDate(0, -5, 0)).Scan(&monthlyStats) // Son 6 ay

	c.JSON(http.StatusOK, gin.H{
		"doctors":           docStats,
		"top_treatments":    trtStats,
		"monthly_revenue":   monthlyStats,
		"total_revenue_ytd": 150000.0, // Örnek YTD
	})
}
