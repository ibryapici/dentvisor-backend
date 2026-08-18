package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TreatmentPlanHandler struct{}

func NewTreatmentPlanHandler() *TreatmentPlanHandler {
	return &TreatmentPlanHandler{}
}

type CreatePlanItemRequest struct {
	ToothNumber   *int    `json:"tooth_number"`
	TreatmentID   *string `json:"treatment_id"`
	TreatmentName string  `json:"treatment_name" binding:"required"`
	UnitPrice     float64 `json:"unit_price" binding:"required"`
	Quantity      int     `json:"quantity"`
	Notes         string  `json:"notes"`
}

type CreateTreatmentPlanRequest struct {
	DoctorID       *string                 `json:"doctor_id"`
	Title          string                  `json:"title" binding:"required"`
	DiscountAmount float64                 `json:"discount_amount"`
	DiscountRate   float64                 `json:"discount_rate"`
	Notes          string                  `json:"notes"`
	Items          []CreatePlanItemRequest `json:"items" binding:"required,min=1"`
}

// GET /api/protected/patients/:id/treatment-plans
func (h *TreatmentPlanHandler) GetPatientTreatmentPlans(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	patientID := c.Param("id")

	var plans []models.TreatmentPlan
	err := database.DB.
		Preload("Doctor.User").
		Preload("Items").
		Where("patient_id = ? AND clinic_id = ?", patientID, clinicID).
		Order("created_at desc").
		Find(&plans).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedavi planları yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, plans)
}

// POST /api/protected/patients/:id/treatment-plans
func (h *TreatmentPlanHandler) CreateTreatmentPlan(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	patientID := c.Param("id")

	// Hastanın bu kliniğe ait olduğunu doğrula
	var patient models.Patient
	if err := database.DB.First(&patient, "id = ? AND clinic_id = ?", patientID, clinicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hasta bulunamadı"})
		return
	}

	var req CreateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri: " + err.Error()})
		return
	}

	// Hesaplamaları yap
	var subtotal float64
	var planItems []models.TreatmentPlanItem

	for _, item := range req.Items {
		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}
		itemTotal := item.UnitPrice * float64(qty)
		subtotal += itemTotal

		planItems = append(planItems, models.TreatmentPlanItem{
			ToothNumber:   item.ToothNumber,
			TreatmentID:   item.TreatmentID,
			TreatmentName: item.TreatmentName,
			UnitPrice:     item.UnitPrice,
			Quantity:      qty,
			TotalPrice:    itemTotal,
			Notes:         item.Notes,
			Status:        "planned",
		})
	}

	discountAmount := req.DiscountAmount
	if req.DiscountRate > 0 {
		discountAmount = (subtotal * req.DiscountRate) / 100
	}
	totalAmount := subtotal - discountAmount
	if totalAmount < 0 {
		totalAmount = 0
	}

	plan := models.TreatmentPlan{
		ClinicID:       clinicID,
		PatientID:      patientID,
		DoctorID:       req.DoctorID,
		Title:          req.Title,
		Status:         "presented", // varsayılan: sunuldu
		Subtotal:       subtotal,
		DiscountAmount: discountAmount,
		DiscountRate:   req.DiscountRate,
		TotalAmount:    totalAmount,
		Notes:          req.Notes,
		Items:          planItems,
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&plan).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedavi planı oluşturulamadı: " + err.Error()})
		return
	}

	// Oluşturulan planı ilişkileriyle getir
	database.DB.Preload("Doctor.User").Preload("Items").First(&plan, "id = ?", plan.ID)
	c.JSON(http.StatusCreated, plan)
}

// GET /api/protected/patients/:id/treatment-plans/:planId
func (h *TreatmentPlanHandler) GetTreatmentPlanDetail(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	planID := c.Param("planId")

	var plan models.TreatmentPlan
	err := database.DB.
		Preload("Patient").
		Preload("Doctor.User").
		Preload("Items").
		First(&plan, "id = ? AND clinic_id = ?", planID, clinicID).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tedavi planı bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

type UpdatePlanStatusRequest struct {
	Status string `json:"status" binding:"required"` // draft, presented, accepted, rejected, completed
}

// PUT /api/protected/patients/:id/treatment-plans/:planId/status
func (h *TreatmentPlanHandler) UpdateTreatmentPlanStatus(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	planID := c.Param("planId")

	var req UpdatePlanStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz durum verisi"})
		return
	}

	if err := database.DB.Model(&models.TreatmentPlan{}).
		Where("id = ? AND clinic_id = ?", planID, clinicID).
		Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Durum güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tedavi planı durumu güncellendi"})
}

// DELETE /api/protected/patients/:id/treatment-plans/:planId
func (h *TreatmentPlanHandler) DeleteTreatmentPlan(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	planID := c.Param("planId")

	if err := database.DB.Where("id = ? AND clinic_id = ?", planID, clinicID).Delete(&models.TreatmentPlan{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedavi planı silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tedavi planı başarıyla silindi"})
}
