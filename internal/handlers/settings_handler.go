package handlers

import (
	"net/http"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	if err := database.DB.Preload("City").Preload("District").First(&clinic, "id = ?", clinicID).Error; err != nil {
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

	var req struct {
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Address      string `json:"address"`
		WorkingHours string `json:"working_hours"`
		AboutText    string `json:"about_text"`
		CityID       *uint  `json:"city_id"`
		DistrictID   *uint  `json:"district_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz form verisi"})
		return
	}

	updates := map[string]interface{}{
		"name":          req.Name,
		"phone":         req.Phone,
		"address":       req.Address,
		"working_hours": req.WorkingHours,
		"about_text":    req.AboutText,
		"city_id":       req.CityID,
		"district_id":   req.DistrictID,
	}

	if err := database.DB.Model(&models.Clinic{}).Where("id = ?", clinicID).Updates(updates).Error; err != nil {
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
	if err := database.DB.Preload("User").Where("clinic_id = ?", clinicID).Order("created_at asc").Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hekimler listelenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, doctors)
}

func (h *SettingsHandler) AddDoctor(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var req struct {
		FirstName      string  `json:"first_name" binding:"required"`
		LastName       string  `json:"last_name" binding:"required"`
		Email          string  `json:"email" binding:"required"`
		Phone          string  `json:"phone"`
		Title          string  `json:"title"`
		Specialty      string  `json:"specialty"`
		CommissionRate float64 `json:"commission_rate"`
		BaseSalary     float64 `json:"base_salary"`
		Password       string  `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen gerekli hekim bilgilerini doldurun"})
		return
	}

	// Set default password if not provided
	pass := req.Password
	if pass == "" {
		pass = "123456"
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Şifre oluşturulamadı"})
		return
	}

	// Check if user with email already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu e-posta adresiyle kayıtlı bir kullanıcı zaten mevcut"})
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Role:         "doctor",
		ClinicID:     &clinicID,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hekim kullanıcısı oluşturulamadı"})
		return
	}

	title := req.Title
	if title == "" {
		title = "Dt."
	}
	specialty := req.Specialty
	if specialty == "" {
		specialty = "Genel Diş Hekimliği"
	}
	commRate := req.CommissionRate
	if commRate <= 0 {
		commRate = 35.0
	}

	doctor := models.Doctor{
		UserID:         user.ID,
		ClinicID:       clinicID,
		Title:          title,
		Specialty:      specialty,
		CommissionRate: commRate,
		BaseSalary:     req.BaseSalary,
	}

	if err := database.DB.Create(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hekim kaydı oluşturulamadı"})
		return
	}

	database.DB.Preload("User").First(&doctor, "id = ?", doctor.ID)
	c.JSON(http.StatusCreated, doctor)
}

func (h *SettingsHandler) UpdateDoctor(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	doctorID := c.Param("id")
	var doctor models.Doctor
	if err := database.DB.Preload("User").Where("id = ? AND clinic_id = ?", doctorID, clinicID).First(&doctor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hekim bulunamadı"})
		return
	}

	var req struct {
		FirstName      string  `json:"first_name"`
		LastName       string  `json:"last_name"`
		Email          string  `json:"email"`
		Phone          string  `json:"phone"`
		Title          string  `json:"title"`
		Specialty      string  `json:"specialty"`
		CommissionRate float64 `json:"commission_rate"`
		BaseSalary     float64 `json:"base_salary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	// Update Doctor details
	doctorUpdates := map[string]interface{}{}
	if req.Title != "" {
		doctorUpdates["title"] = req.Title
	}
	if req.Specialty != "" {
		doctorUpdates["specialty"] = req.Specialty
	}
	if req.CommissionRate > 0 {
		doctorUpdates["commission_rate"] = req.CommissionRate
	}
	if req.BaseSalary >= 0 {
		doctorUpdates["base_salary"] = req.BaseSalary
	}

	if len(doctorUpdates) > 0 {
		database.DB.Model(&doctor).Updates(doctorUpdates)
	}

	// Update User details
	userUpdates := map[string]interface{}{}
	if req.FirstName != "" {
		userUpdates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		userUpdates["last_name"] = req.LastName
	}
	if req.Phone != "" {
		userUpdates["phone"] = req.Phone
	}
	if req.Email != "" && req.Email != doctor.User.Email {
		var existingUser models.User
		if err := database.DB.Where("email = ? AND id != ?", req.Email, doctor.UserID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bu e-posta adresi başka bir kullanıcı tarafından kullanılıyor"})
			return
		}
		userUpdates["email"] = req.Email
	}

	if len(userUpdates) > 0 {
		database.DB.Model(&models.User{}).Where("id = ?", doctor.UserID).Updates(userUpdates)
	}

	database.DB.Preload("User").First(&doctor, "id = ?", doctor.ID)
	c.JSON(http.StatusOK, doctor)
}

func (h *SettingsHandler) DeleteDoctor(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	doctorID := c.Param("id")
	var doctor models.Doctor
	if err := database.DB.Where("id = ? AND clinic_id = ?", doctorID, clinicID).First(&doctor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hekim bulunamadı"})
		return
	}

	if err := database.DB.Delete(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hekim silinirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hekim başarıyla silindi"})
}

// --- TREATMENTS ---
func (h *SettingsHandler) GetTreatments(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var treatments []models.Treatment
	if err := database.DB.Where("clinic_id = ?", clinicID).Order("name asc").Find(&treatments).Error; err != nil {
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
	if req.DefaultDuration <= 0 {
		req.DefaultDuration = 30
	}

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedavi eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *SettingsHandler) UpdateTreatment(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	id := c.Param("id")
	var treatment models.Treatment
	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).First(&treatment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tedavi bulunamadı"})
		return
	}

	var req struct {
		Name            string  `json:"name"`
		DefaultDuration int     `json:"default_duration"`
		DefaultPrice    float64 `json:"default_price"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz form verisi"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.DefaultDuration > 0 {
		updates["default_duration"] = req.DefaultDuration
	}
	if req.DefaultPrice >= 0 {
		updates["default_price"] = req.DefaultPrice
	}

	if err := database.DB.Model(&treatment).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedavi güncellenirken hata oluştu"})
		return
	}

	database.DB.First(&treatment, "id = ?", id)
	c.JSON(http.StatusOK, treatment)
}

func (h *SettingsHandler) DeleteTreatment(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	id := c.Param("id")
	var treatment models.Treatment
	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).First(&treatment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tedavi bulunamadı"})
		return
	}

	if err := database.DB.Delete(&treatment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tedavi silinirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tedavi başarıyla silindi"})
}

// --- CHAIRS ---
func (h *SettingsHandler) GetChairs(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	var chairs []models.Chair
	if err := database.DB.Where("clinic_id = ?", clinicID).Order("name asc").Find(&chairs).Error; err != nil {
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
	if req.Status == "" {
		req.Status = "active"
	}

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ünit eklenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *SettingsHandler) UpdateChair(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	id := c.Param("id")
	var chair models.Chair
	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).First(&chair).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ünit bulunamadı"})
		return
	}

	var req struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz form verisi"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := database.DB.Model(&chair).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ünit güncellenirken hata oluştu"})
		return
	}

	database.DB.First(&chair, "id = ?", id)
	c.JSON(http.StatusOK, chair)
}

func (h *SettingsHandler) DeleteChair(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	id := c.Param("id")
	var chair models.Chair
	if err := database.DB.Where("id = ? AND clinic_id = ?", id, clinicID).First(&chair).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ünit bulunamadı"})
		return
	}

	if err := database.DB.Delete(&chair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ünit silinirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ünit başarıyla silindi"})
}
