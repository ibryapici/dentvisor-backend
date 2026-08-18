package handlers

import (
	"net/http"
	"time"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type LabHandler struct{}

func NewLabHandler() *LabHandler {
	return &LabHandler{}
}

// --- LABORATORIES ---

type CreateLabRequest struct {
	Name          string `json:"name" binding:"required"`
	Phone         string `json:"phone"`
	ContactPerson string `json:"contact_person"`
	Address       string `json:"address"`
	Notes         string `json:"notes"`
}

// GET /api/protected/laboratories
func (h *LabHandler) GetLaboratories(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var labs []models.Laboratory
	if err := database.DB.Where("clinic_id = ?", clinicID).Order("name asc").Find(&labs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Laboratuvarlar yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, labs)
}

// POST /api/protected/laboratories
func (h *LabHandler) CreateLaboratory(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var req CreateLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Laboratuvar adı zorunludur"})
		return
	}

	lab := models.Laboratory{
		ClinicID:      clinicID,
		Name:          req.Name,
		Phone:         req.Phone,
		ContactPerson: req.ContactPerson,
		Address:       req.Address,
		Notes:         req.Notes,
	}

	if err := database.DB.Create(&lab).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Laboratuvar kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, lab)
}

// PUT /api/protected/laboratories/:id
func (h *LabHandler) UpdateLaboratory(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	id := c.Param("id")
	var lab models.Laboratory
	if err := database.DB.First(&lab, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laboratuvar bulunamadı"})
		return
	}

	var req CreateLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	lab.Name = req.Name
	lab.Phone = req.Phone
	lab.ContactPerson = req.ContactPerson
	lab.Address = req.Address
	lab.Notes = req.Notes

	if err := database.DB.Save(&lab).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, lab)
}

// DELETE /api/protected/laboratories/:id
func (h *LabHandler) DeleteLaboratory(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	id := c.Param("id")

	// Bağlı iş var mı kontrol et
	var orderCount int64
	database.DB.Model(&models.LabOrder{}).Where("laboratory_id = ?", id).Count(&orderCount)
	if orderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu laboratuvara ait kayıtlı protez/iş kayıtları bulunduğu için silemezsiniz."})
		return
	}

	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).Delete(&models.Laboratory{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Laboratuvar silindi"})
}

// --- LAB ORDERS ---

type CreateLabOrderRequest struct {
	PatientID    string     `json:"patient_id" binding:"required"`
	DoctorID     *string    `json:"doctor_id"`
	LaboratoryID string     `json:"laboratory_id" binding:"required"`
	WorkType     string     `json:"work_type" binding:"required"`
	ToothNumbers string     `json:"tooth_numbers"`
	ShadeColor   string     `json:"shade_color"`
	DueDate      *time.Time `json:"due_date"`
	Cost         float64    `json:"cost"`
	Notes        string     `json:"notes"`
}

// GET /api/protected/lab-orders
func (h *LabHandler) GetLabOrders(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	status := c.Query("status")
	labID := c.Query("laboratory_id")
	patientID := c.Query("patient_id")

	query := database.DB.Model(&models.LabOrder{}).
		Preload("Patient").
		Preload("Doctor.User").
		Preload("Laboratory").
		Where("clinic_id = ?", clinicID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if labID != "" {
		query = query.Where("laboratory_id = ?", labID)
	}
	if patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	var orders []models.LabOrder
	if err := query.Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İşler yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// POST /api/protected/lab-orders
func (h *LabHandler) CreateLabOrder(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var req CreateLabOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen hasta, laboratuvar ve iş türü alanlarını eksiksiz doldurunuz."})
		return
	}

	order := models.LabOrder{
		ClinicID:     clinicID,
		PatientID:    req.PatientID,
		DoctorID:     req.DoctorID,
		LaboratoryID: req.LaboratoryID,
		WorkType:     req.WorkType,
		ToothNumbers: req.ToothNumbers,
		ShadeColor:   req.ShadeColor,
		Status:       "sent", // Ölçü Gönderildi
		SentDate:     time.Now(),
		DueDate:      req.DueDate,
		Cost:         req.Cost,
		Notes:        req.Notes,
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Laboratuvar işi oluşturulamadı"})
		return
	}

	database.DB.Preload("Patient").Preload("Doctor.User").Preload("Laboratory").First(&order, "id = ?", order.ID)
	c.JSON(http.StatusCreated, order)
}

type UpdateLabOrderStatusRequest struct {
	Status string `json:"status" binding:"required"` // sent, in_lab, try_in, completed, cancelled
}

// PUT /api/protected/lab-orders/:id/status
func (h *LabHandler) UpdateLabOrderStatus(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	id := c.Param("id")
	var req UpdateLabOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz durum bilgisi"})
		return
	}

	updates := map[string]interface{}{
		"status": req.Status,
	}
	if req.Status == "completed" {
		now := time.Now()
		updates["completed_date"] = &now
	}

	if err := database.DB.Model(&models.LabOrder{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Durum güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "İş durumu güncellendi"})
}

// PUT /api/protected/lab-orders/:id
func (h *LabHandler) UpdateLabOrder(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	id := c.Param("id")
	var order models.LabOrder
	if err := database.DB.First(&order, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "İş kaydı bulunamadı"})
		return
	}

	var req struct {
		LaboratoryID string     `json:"laboratory_id"`
		WorkType     string     `json:"work_type"`
		ToothNumbers string     `json:"tooth_numbers"`
		ShadeColor   string     `json:"shade_color"`
		DueDate      *time.Time `json:"due_date"`
		Cost         float64    `json:"cost"`
		IsPaid       bool       `json:"is_paid"`
		Notes        string     `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	if req.LaboratoryID != "" {
		order.LaboratoryID = req.LaboratoryID
	}
	if req.WorkType != "" {
		order.WorkType = req.WorkType
	}
	order.ToothNumbers = req.ToothNumbers
	order.ShadeColor = req.ShadeColor
	order.DueDate = req.DueDate
	order.Cost = req.Cost
	order.IsPaid = req.IsPaid
	order.Notes = req.Notes

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kayıt güncellenemedi"})
		return
	}

	database.DB.Preload("Patient").Preload("Doctor.User").Preload("Laboratory").First(&order, "id = ?", order.ID)
	c.JSON(http.StatusOK, order)
}

// DELETE /api/protected/lab-orders/:id
func (h *LabHandler) DeleteLabOrder(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	id := c.Param("id")
	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).Delete(&models.LabOrder{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Laboratuvar işi başarıyla silindi"})
}
