package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// GET /api/protected/sistem/clinics
func (h *SystemHandler) GetClinics(c *gin.Context) {
	var clinics []models.Clinic
	
	// We might want to omit some heavy relational data or include basic counts.
	// For now, let's preload users or just return clinics.
	if err := database.DB.Preload("Users").Order("created_at desc").Find(&clinics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Klinikler yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, clinics)
}

type UpdateClinicStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// PUT /api/protected/sistem/clinics/:id/status
func (h *SystemHandler) UpdateClinicStatus(c *gin.Context) {
	id := c.Param("id")
	
	var req UpdateClinicStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	result := database.DB.Model(&models.Clinic{}).Where("id = ?", id).Update("is_active", req.IsActive)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Durum güncellenemedi"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Klinik bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Klinik durumu güncellendi"})
}
