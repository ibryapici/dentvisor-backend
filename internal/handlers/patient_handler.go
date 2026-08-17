package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct{}

func NewPatientHandler() *PatientHandler {
	return &PatientHandler{}
}

// GET /api/protected/patients
func (h *PatientHandler) GetPatients(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var patients []models.Patient
	if err := database.DB.Where("clinic_id = ?", clinicID).Find(&patients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hastalar listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, patients)
}

// POST /api/protected/patients
func (h *PatientHandler) AddPatient(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req models.Patient
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	req.ClinicID = clinicID

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hasta eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, req)
}

// GET /api/protected/patients/:id
func (h *PatientHandler) GetPatientDetail(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	patientID := c.Param("id")
	var patient models.Patient

	// Preload DentalRecords, Payments, Appointments
	err := database.DB.Preload("DentalRecords").
		Preload("DentalRecords.Treatment").
		Preload("DentalRecords.Doctor").
		Preload("Payments").
		Preload("Appointments").
		Where("id = ? AND clinic_id = ?", patientID, clinicID).
		First(&patient).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hasta bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// POST /api/protected/patients/:id/dental-records
func (h *PatientHandler) AddDentalRecord(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	patientID := c.Param("id")
	var req models.DentalRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	// Verify Patient belongs to Clinic
	var count int64
	database.DB.Model(&models.Patient{}).Where("id = ? AND clinic_id = ?", patientID, clinicID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Yetkisiz erişim"})
		return
	}

	req.PatientID = patientID

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kayıt eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, req)
}
