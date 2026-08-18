package handlers

import (
	"fmt"
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"dentvisor-backend/internal/services"
	"dentvisor-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{
	authService *services.AuthService
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		authService: services.NewAuthService(),
	}
}

// GET /api/protected/sistem/dashboard
func (h *SystemHandler) GetDashboardStats(c *gin.Context) {
	var totalClinics int64
	var activeClinics int64
	var totalUsers int64
	var totalPatients int64

	database.DB.Model(&models.Clinic{}).Count(&totalClinics)
	database.DB.Model(&models.Clinic{}).Where("is_active = ?", true).Count(&activeClinics)
	database.DB.Model(&models.User{}).Count(&totalUsers)
	database.DB.Model(&models.Patient{}).Count(&totalPatients)

	c.JSON(http.StatusOK, gin.H{
		"total_clinics":  totalClinics,
		"active_clinics": activeClinics,
		"total_users":    totalUsers,
		"total_patients": totalPatients,
	})
}

// GET /api/protected/sistem/clinics
func (h *SystemHandler) GetClinics(c *gin.Context) {
	page, limit, search := utils.GetPaginationParams(c)

	query := database.DB.Model(&models.Clinic{})
	
	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where("name ILIKE ? OR phone ILIKE ?", searchLike, searchLike)
	}
	
	query = query.Order("created_at desc")

	var clinics []models.Clinic
	result, err := utils.Paginate(query, page, limit, &clinics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Klinikler yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, result)
}

type UpdateClinicRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// PUT /api/protected/sistem/clinics/:id
func (h *SystemHandler) UpdateClinic(c *gin.Context) {
	id := c.Param("id")
	
	var req UpdateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	result := database.DB.Model(&models.Clinic{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":  req.Name,
		"phone": req.Phone,
	})
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Klinik güncellenemedi"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Klinik bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Klinik başarıyla güncellendi"})
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

// POST /api/protected/sistem/impersonate
func (h *SystemHandler) Impersonate(c *gin.Context) {
	var req struct {
		ClinicID string `json:"clinic_id"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	token, user, err := h.authService.ImpersonateClinic(req.ClinicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
		"message": "Kliniğe bağlanıldı",
	})
}

// GET /api/protected/sistem/patients
func (h *SystemHandler) GetAllPatients(c *gin.Context) {
	page, limit, search := utils.GetPaginationParams(c)

	query := database.DB.Model(&models.Patient{}).Preload("Clinic")
	
	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR phone ILIKE ? OR tc_no ILIKE ?", searchLike, searchLike, searchLike, searchLike)
	}
	
	query = query.Order("created_at desc")

	var patients []models.Patient
	result, err := utils.Paginate(query, page, limit, &patients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hastalar yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /api/protected/sistem/appointments
func (h *SystemHandler) GetAllAppointments(c *gin.Context) {
	page, limit, search := utils.GetPaginationParams(c)

	query := database.DB.Model(&models.Appointment{}).Preload("Patient").Preload("Doctor").Preload("Clinic")
	
	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Joins("JOIN patients on patients.id = appointments.patient_id").
			Where("patients.first_name ILIKE ? OR patients.last_name ILIKE ?", searchLike, searchLike)
	}
	
	query = query.Order("start_time desc")

	var appointments []models.Appointment
	result, err := utils.Paginate(query, page, limit, &appointments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Randevular yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// --- LOKASYON YÖNETİMİ ---

func (h *SystemHandler) GetSystemCities(c *gin.Context) {
	var cities []models.City
	if err := database.DB.Order("name asc").Find(&cities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İller yüklenemedi"})
		return
	}
	c.JSON(http.StatusOK, cities)
}

func (h *SystemHandler) CreateCity(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}
	city := models.City{Name: req.Name}
	if err := database.DB.Create(&city).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İl oluşturulamadı"})
		return
	}
	c.JSON(http.StatusOK, city)
}

func (h *SystemHandler) UpdateCity(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}
	if err := database.DB.Model(&models.City{}).Where("id = ?", id).Update("name", req.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İl güncellenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "İl güncellendi"})
}

func (h *SystemHandler) DeleteCity(c *gin.Context) {
	id := c.Param("id")
	var count int64
	database.DB.Model(&models.Clinic{}).Where("city_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu ile bağlı klinikler olduğu için silemezsiniz."})
		return
	}
	
	// Delete dependent districts first
	database.DB.Where("city_id = ?", id).Delete(&models.District{})
	
	if err := database.DB.Where("id = ?", id).Delete(&models.City{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İl silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "İl silindi"})
}

func (h *SystemHandler) GetSystemDistricts(c *gin.Context) {
	cityId := c.Param("cityId")
	var districts []models.District
	if err := database.DB.Where("city_id = ?", cityId).Order("name asc").Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlçeler yüklenemedi"})
		return
	}
	c.JSON(http.StatusOK, districts)
}

func (h *SystemHandler) CreateDistrict(c *gin.Context) {
	cityId := c.Param("cityId")
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}
	
	// Parse cityId as uint
	var cityIdUint uint
	fmt.Sscanf(cityId, "%d", &cityIdUint)
	
	district := models.District{Name: req.Name, CityID: cityIdUint}
	if err := database.DB.Create(&district).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlçe oluşturulamadı"})
		return
	}
	c.JSON(http.StatusOK, district)
}

func (h *SystemHandler) UpdateDistrict(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}
	if err := database.DB.Model(&models.District{}).Where("id = ?", id).Update("name", req.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlçe güncellenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "İlçe güncellendi"})
}

func (h *SystemHandler) DeleteDistrict(c *gin.Context) {
	id := c.Param("id")
	var count int64
	database.DB.Model(&models.Clinic{}).Where("district_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu ilçeye bağlı klinikler olduğu için silemezsiniz."})
		return
	}
	if err := database.DB.Where("id = ?", id).Delete(&models.District{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlçe silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "İlçe silindi"})
}
