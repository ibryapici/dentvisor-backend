package handlers

import (
	"net/http"
	"time"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct{}

func NewAppointmentHandler() *AppointmentHandler {
	return &AppointmentHandler{}
}

// GET /api/protected/appointments
func (h *AppointmentHandler) GetAppointments(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var appointments []models.Appointment
	// Preload relations so calendar can show names
	if err := database.DB.Preload("Patient").Preload("Doctor").Preload("Doctor.User").Preload("Treatment").Preload("Chair").
		Where("clinic_id = ?", clinicID).Find(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Randevular listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, appointments)
}

// POST /api/protected/appointments
func (h *AppointmentHandler) AddAppointment(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req struct {
		PatientID   string    `json:"patient_id" binding:"required"`
		DoctorID    string    `json:"doctor_id" binding:"required"`
		TreatmentID string    `json:"treatment_id" binding:"required"`
		ChairID     string    `json:"chair_id" binding:"required"`
		StartTime   time.Time `json:"start_time" binding:"required"`
		EndTime     time.Time `json:"end_time" binding:"required"`
		Notes       string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri: " + err.Error()})
		return
	}

	// 1. Çakışma (Overlap) Denetimi
	// Aynı koltukta (ChairID) veya aynı doktorda (DoctorID) bu saatler arasında başka randevu var mı?
	var overlapCount int64
	database.DB.Model(&models.Appointment{}).Where(
		"clinic_id = ? AND (doctor_id = ? OR chair_id = ?) AND ((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?)) AND status != 'cancelled'",
		clinicID, req.DoctorID, req.ChairID,
		req.EndTime, req.StartTime, // overlap condition 1
		req.EndTime, req.StartTime, // overlap condition 2
		req.StartTime, req.EndTime, // overlap condition 3
	).Count(&overlapCount)

	if overlapCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Seçilen tarih ve saatte bu doktorun veya koltuğun başka bir randevusu mevcut."})
		return
	}

	appointment := models.Appointment{
		ClinicID:    clinicID,
		PatientID:   req.PatientID,
		DoctorID:    req.DoctorID,
		TreatmentID: req.TreatmentID,
		ChairID:     req.ChairID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Notes:       req.Notes,
		Status:      "scheduled",
	}

	if err := database.DB.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Randevu eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, appointment)
}

// PUT /api/protected/appointments/:id
func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	id := c.Param("id")
	var appointment models.Appointment
	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).First(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Randevu bulunamadı"})
		return
	}

	var req struct {
		DoctorID    string     `json:"doctor_id"`
		TreatmentID string     `json:"treatment_id"`
		ChairID     string     `json:"chair_id"`
		StartTime   *time.Time `json:"start_time"`
		EndTime     *time.Time `json:"end_time"`
		Notes       *string    `json:"notes"`
		Status      string     `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri: " + err.Error()})
		return
	}

	updates := map[string]interface{}{}

	newDoctorID := appointment.DoctorID
	if req.DoctorID != "" {
		newDoctorID = req.DoctorID
		updates["doctor_id"] = req.DoctorID
	}
	newChairID := appointment.ChairID
	if req.ChairID != "" {
		newChairID = req.ChairID
		updates["chair_id"] = req.ChairID
	}
	if req.TreatmentID != "" {
		updates["treatment_id"] = req.TreatmentID
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	newStartTime := appointment.StartTime
	if req.StartTime != nil && !req.StartTime.IsZero() {
		newStartTime = *req.StartTime
		updates["start_time"] = *req.StartTime
	}
	newEndTime := appointment.EndTime
	if req.EndTime != nil && !req.EndTime.IsZero() {
		newEndTime = *req.EndTime
		updates["end_time"] = *req.EndTime
	}

	// Check overlap excluding current appointment
	if updates["status"] != "cancelled" {
		var overlapCount int64
		database.DB.Model(&models.Appointment{}).Where(
			"id != ? AND clinic_id = ? AND (doctor_id = ? OR chair_id = ?) AND ((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?)) AND status != 'cancelled'",
			id, clinicID, newDoctorID, newChairID,
			newEndTime, newStartTime,
			newEndTime, newStartTime,
			newStartTime, newEndTime,
		).Count(&overlapCount)

		if overlapCount > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Seçilen tarih ve saatte bu doktorun veya koltuğun başka bir randevusu mevcut."})
			return
		}
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&appointment).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Randevu güncellenirken hata oluştu"})
			return
		}
	}

	database.DB.Preload("Patient").Preload("Doctor").Preload("Doctor.User").Preload("Treatment").Preload("Chair").First(&appointment, "id = ?", id)
	c.JSON(http.StatusOK, appointment)
}

// DELETE /api/protected/appointments/:id
func (h *AppointmentHandler) DeleteAppointment(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	id := c.Param("id")
	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).Delete(&models.Appointment{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Randevu silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Randevu başarıyla silindi"})
}
