package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type ContentHandler struct{}

func NewContentHandler() *ContentHandler {
	return &ContentHandler{}
}

// --- REVIEWS (YORUMLAR) ---
func (h *ContentHandler) GetReviews(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var reviews []models.Review
	if err := database.DB.Preload("Patient").Where("clinic_id = ?", clinicID).Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yorumlar listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (h *ContentHandler) UpdateReviewStatus(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	reviewID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"` // approved, rejected
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz durum"})
		return
	}

	if err := database.DB.Model(&models.Review{}).Where("id = ? AND clinic_id = ?", reviewID, clinicID).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yorum güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Yorum durumu güncellendi"})
}

// --- ARTICLES (YAZILAR/BLOG) ---
func (h *ContentHandler) GetArticles(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var articles []models.Article
	if err := database.DB.Where("clinic_id = ?", clinicID).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yazılar listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

func (h *ContentHandler) AddArticle(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req models.Article
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}
	
	req.ClinicID = clinicID

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yazı eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, req)
}
