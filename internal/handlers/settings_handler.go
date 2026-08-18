package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// Helper to get clinicID from token claims
func getClinicID(c *gin.Context) (string, bool) {
	// First check context (set by middleware from claims)
	if clinicID, exists := c.Get("clinic_id"); exists {
		if id, ok := clinicID.(string); ok {
			return id, true
		}
	}
	
	// Fallback to query database using userID if not in claims
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	
	// Fallback to query database if not in claims
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return "", false
	}
	
	if user.ClinicID == nil {
		return "", false
	}
	
	return *user.ClinicID, true
}

// --- CLINIC PROFILE ---
func (h *SettingsHandler) GetClinicProfile(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var clinic models.Clinic
	if err := database.DB.First(&clinic, "id = ?", clinicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Klinik bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, clinic)
}

func (h *SettingsHandler) UpdateClinicProfile(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req models.Clinic
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz form verisi"})
		return
	}

	if err := database.DB.Model(&models.Clinic{}).Where("id = ?", clinicID).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Klinik güncellenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Klinik profili başarıyla güncellendi"})
}

// --- DOCTORS ---
func (h *SettingsHandler) GetDoctors(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var doctors []models.Doctor
	if err := database.DB.Preload("User").Where("clinic_id = ?", clinicID).Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hekimler listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, doctors)
}

// --- TREATMENTS ---
func (h *SettingsHandler) GetTreatments(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var treatments []models.Treatment
	if err := database.DB.Where("clinic_id = ?", clinicID).Find(&treatments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedaviler listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, treatments)
}

func (h *SettingsHandler) AddTreatment(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req models.Treatment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}
	
	req.ClinicID = clinicID

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedavi eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, req)
}

// --- CHAIRS ---
func (h *SettingsHandler) GetChairs(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var chairs []models.Chair
	if err := database.DB.Where("clinic_id = ?", clinicID).Find(&chairs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ünitler listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, chairs)
}

func (h *SettingsHandler) AddChair(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req models.Chair
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}
	
	req.ClinicID = clinicID

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ünit eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, req)
}
