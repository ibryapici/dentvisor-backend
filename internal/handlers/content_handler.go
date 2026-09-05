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

func (h *ContentHandler) UpdateArticle(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	articleID := c.Param("id")
	var article models.Article
	if err := database.DB.Where("id = ? AND clinic_id = ?", articleID, clinicID).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Yazı bulunamadı"})
		return
	}

	var req struct {
		Title     string `json:"title"`
		Slug      string `json:"slug"`
		Content   string `json:"content"`
		CoverURL  string `json:"cover_url"`
		Published *bool  `json:"published"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Slug != "" {
		updates["slug"] = req.Slug
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.CoverURL != "" {
		updates["cover_url"] = req.CoverURL
	}
	if req.Published != nil {
		updates["published"] = *req.Published
	}

	if err := database.DB.Model(&article).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yazı güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, article)
}

func (h *ContentHandler) DeleteArticle(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	articleID := c.Param("id")
	if err := database.DB.Where("id = ? AND clinic_id = ?", articleID, clinicID).Delete(&models.Article{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yazı silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Yazı başarıyla silindi"})
}

