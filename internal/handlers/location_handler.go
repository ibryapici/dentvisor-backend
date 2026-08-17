package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type LocationHandler struct{}

func NewLocationHandler() *LocationHandler {
	return &LocationHandler{}
}

// GET /api/public/cities
func (h *LocationHandler) GetCities(c *gin.Context) {
	var cities []models.City
	if err := database.DB.Order("name ASC").Find(&cities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İller yüklenemedi"})
		return
	}
	c.JSON(http.StatusOK, cities)
}

// GET /api/public/cities/:id/districts
func (h *LocationHandler) GetDistrictsByCity(c *gin.Context) {
	cityID := c.Param("id")
	var districts []models.District
	if err := database.DB.Where("city_id = ?", cityID).Order("name ASC").Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlçeler yüklenemedi"})
		return
	}
	c.JSON(http.StatusOK, districts)
}
