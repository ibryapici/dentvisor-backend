package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct{}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

// GET /api/protected/patients/:id/payments
func (h *PaymentHandler) GetPatientPayments(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	patientID := c.Param("id")

	var payments []models.Payment
	if err := database.DB.Where("clinic_id = ? AND patient_id = ?", clinicID, patientID).Order("created_at desc").Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ödemeler getirilirken bir hata oluştu"})
		return
	}

	// Hastanın toplam borcunu/bakiyesini de hesaplayıp dönmek faydalı olabilir
	// Bakiyeyi basitçe: (Yapılan tedavilerin toplam fiyatı) - (Ödenen miktar) olarak hesaplayabiliriz.
	// Şimdilik sadece ödemeleri dönüyoruz.
	c.JSON(http.StatusOK, payments)
}

// POST /api/protected/patients/:id/payments
func (h *PaymentHandler) AddPayment(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	patientID := c.Param("id")

	var input struct {
		Amount        float64 `json:"amount" binding:"required"`
		PaymentMethod string  `json:"payment_method" binding:"required"`
		Type          string  `json:"type" binding:"required"` // income, refund
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
		return
	}

	payment := models.Payment{
		ClinicID:      clinicID,
		PatientID:     patientID,
		Amount:        input.Amount,
		PaymentMethod: input.PaymentMethod,
		Type:          input.Type,
	}

	if err := database.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ödeme kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, payment)
}
