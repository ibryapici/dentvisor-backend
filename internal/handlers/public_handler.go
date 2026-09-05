package handlers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/internal/utils"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PublicHandler struct{}

func NewPublicHandler() *PublicHandler {
	return &PublicHandler{}
}

type PublicClinicResponse struct {
	models.Clinic
	IsClaimed bool `json:"is_claimed"`
}

// GET /api/public/clinics
func (h *PublicHandler) GetClinics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	search := c.Query("search")
	cityID := c.Query("city_id")
	districtID := c.Query("district_id")

	query := database.DB.Model(&models.Clinic{}).
		Preload("City").
		Preload("District").
		Preload("Treatments").
		Where("is_active = ?", true)

	if search != "" {
		searchLike := "%" + search + "%"
		query = query.Where("name ILIKE ? OR address ILIKE ? OR about_text ILIKE ?", searchLike, searchLike, searchLike)
	}

	if cityID != "" {
		query = query.Where("city_id = ?", cityID)
	}

	if districtID != "" {
		query = query.Where("district_id = ?", districtID)
	}

	query = query.Order("name ASC")

	var clinics []models.Clinic
	result, err := utils.Paginate(query, page, limit, &clinics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Klinikler yüklenemedi"})
		return
	}

	// Claim durumunu tespit edelim (User tablosunda bu kliniğe bağlı doctor/admin var mı?)
	var clinicIDs []string
	for _, cl := range clinics {
		clinicIDs = append(clinicIDs, cl.ID)
	}

	claimedMap := make(map[string]bool)
	if len(clinicIDs) > 0 {
		var claimedClinicIDs []string
		database.DB.Model(&models.User{}).
			Where("clinic_id IN ? AND role IN ?", clinicIDs, []string{"doctor", "admin"}).
			Pluck("DISTINCT clinic_id", &claimedClinicIDs)

		for _, cid := range claimedClinicIDs {
			claimedMap[cid] = true
		}
	}

	var responseClinics []PublicClinicResponse
	for _, cl := range clinics {
		responseClinics = append(responseClinics, PublicClinicResponse{
			Clinic:    cl,
			IsClaimed: claimedMap[cl.ID],
		})
	}

	result.Data = responseClinics
	c.JSON(http.StatusOK, result)
}

// GET /api/public/clinics/:id
func (h *PublicHandler) GetClinicDetail(c *gin.Context) {
	id := c.Param("id")

	var clinic models.Clinic
	err := database.DB.
		Preload("City").
		Preload("District").
		Preload("Doctors.User").
		Preload("Treatments").
		Preload("Articles").
		Preload("Reviews", "status = ?", "approved").
		First(&clinic, "id = ? AND is_active = ?", id, true).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Klinik bulunamadı"})
		return
	}

	var userCount int64
	database.DB.Model(&models.User{}).
		Where("clinic_id = ? AND role IN ?", clinic.ID, []string{"doctor", "admin"}).
		Count(&userCount)

	c.JSON(http.StatusOK, PublicClinicResponse{
		Clinic:    clinic,
		IsClaimed: userCount > 0,
	})
}

type ClaimClinicRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
}

// POST /api/public/clinics/:id/claim
func (h *PublicHandler) ClaimClinic(c *gin.Context) {
	clinicID := c.Param("id")

	var req ClaimClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen tüm alanları geçerli şekilde doldurunuz: " + err.Error()})
		return
	}

	// 1. Klinik var mı ve aktif mi?
	var clinic models.Clinic
	if err := database.DB.First(&clinic, "id = ?", clinicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Klinik bulunamadı"})
		return
	}

	// 2. Zaten sahiplenilmiş mi?
	var existingUsersCount int64
	database.DB.Model(&models.User{}).
		Where("clinic_id = ? AND role IN ?", clinicID, []string{"doctor", "admin"}).
		Count(&existingUsersCount)

	if existingUsersCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Bu klinik zaten doğrulanmış bir yetkili tarafından sahiplenilmiştir. Bilgilerde hata olduğunu düşünüyorsanız destek ile iletişime geçiniz."})
		return
	}

	// 3. E-posta zaten kullanımda mı?
	var emailCheck int64
	database.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&emailCheck)
	if emailCheck > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu e-posta adresi zaten kullanımda."})
		return
	}

	// 4. Şifre hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Şifre işlenirken hata oluştu"})
		return
	}

	var createdUser models.User

	// 5. Transaction ile kullanıcı ve doktor profilini oluştur
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		user := models.User{
			ClinicID:     &clinic.ID,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			Role:         "doctor", // Sahiplenen hekim/yönetici
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Phone:        req.Phone,
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		createdUser = user

		doctor := models.Doctor{
			UserID:   user.ID,
			ClinicID: clinic.ID,
			Title:    "Dt.",
		}
		if err := tx.Create(&doctor).Error; err != nil {
			return err
		}

		// Kliniğin telefon veya bilgilerini güncelle
		if clinic.Phone == "" && req.Phone != "" {
			tx.Model(&clinic).Update("phone", req.Phone)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Klinik sahiplenme işlemi gerçekleştirilemedi: " + err.Error()})
		return
	}

	// 6. Otomatik JWT Token üret
	claims := jwt.MapClaims{
		"sub":       createdUser.ID,
		"role":      createdUser.Role,
		"clinic_id": clinic.ID,
		"exp":       time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Oturum token'ı oluşturulamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tebrikler! Klinik başarıyla sahiplenildi ve hesabınız oluşturuldu.",
		"token":   tokenString,
		"user":    createdUser,
		"clinic":  clinic,
	})
}

// GET /api/public/clinics/:id/slots
func (h *PublicHandler) GetSlots(c *gin.Context) {
	clinicID := c.Param("id")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	durationStr := c.DefaultQuery("duration", "30")
	doctorID := c.Query("doctor_id")

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		duration = 30
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz tarih formatı (YYYY-MM-DD bekleniyor)"})
		return
	}

	// 1. Kliniğin hekim ve koltuklarını al
	var doctors []models.Doctor
	database.DB.Where("clinic_id = ?", clinicID).Find(&doctors)
	if len(doctors) == 0 {
		c.JSON(http.StatusOK, []string{})
		return
	}

	var chairs []models.Chair
	database.DB.Where("clinic_id = ? AND status != 'maintenance'", clinicID).Find(&chairs)
	if len(chairs) == 0 {
		// Koltuk yoksa varsayılan sanal 1 koltuk var say
		chairs = append(chairs, models.Chair{Name: "Ünit-1"})
	}

	// 2. O güne ait mevcut randevuları çek
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	dayEnd := dayStart.Add(24 * time.Hour)

	var appts []models.Appointment
	database.DB.Where("clinic_id = ? AND start_time >= ? AND start_time < ? AND status != 'cancelled'", clinicID, dayStart, dayEnd).Find(&appts)

	// 3. Çalışma saatleri: 09:00 - 18:30 arası slotlar
	var availableSlots []string
	slotStart := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.Local)
	clinicClosing := time.Date(date.Year(), date.Month(), date.Day(), 19, 0, 0, 0, time.Local)

	now := time.Now()

	for slotStart.Add(time.Duration(duration)*time.Minute).Before(clinicClosing) || slotStart.Add(time.Duration(duration)*time.Minute).Equal(clinicClosing) {
		slotEnd := slotStart.Add(time.Duration(duration) * time.Minute)

		// Geçmiş saatleri gösterme
		if slotStart.Before(now) {
			slotStart = slotStart.Add(30 * time.Minute)
			continue
		}

		// Bu slottaki çakışan randevuları say
		occupiedChairs := make(map[string]bool)
		occupiedDoctors := make(map[string]bool)

		for _, a := range appts {
			if (slotStart.Before(a.EndTime) && slotEnd.After(a.StartTime)) ||
				(slotStart.Equal(a.StartTime) || slotEnd.Equal(a.EndTime)) {
				if a.ChairID != "" {
					occupiedChairs[a.ChairID] = true
				}
				if a.DoctorID != "" {
					occupiedDoctors[a.DoctorID] = true
				}
			}
		}

		// Boşta koltuk var mı?
		freeChairCount := len(chairs) - len(occupiedChairs)

		// Hekim uygun mu?
		hasFreeDoctor := false
		if doctorID != "" {
			if !occupiedDoctors[doctorID] {
				hasFreeDoctor = true
			}
		} else {
			if len(doctors) > len(occupiedDoctors) {
				hasFreeDoctor = true
			}
		}

		if freeChairCount > 0 && hasFreeDoctor {
			availableSlots = append(availableSlots, slotStart.Format("15:04"))
		}

		slotStart = slotStart.Add(30 * time.Minute)
	}

	c.JSON(http.StatusOK, availableSlots)
}

type PublicAppointmentRequest struct {
	FirstName   string    `json:"first_name" binding:"required"`
	LastName    string    `json:"last_name" binding:"required"`
	Phone       string    `json:"phone" binding:"required"`
	Email       string    `json:"email"`
	TreatmentID string    `json:"treatment_id" binding:"required"`
	DoctorID    string    `json:"doctor_id"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	Notes       string    `json:"notes"`
}

// POST /api/public/clinics/:id/appointments
func (h *PublicHandler) CreatePublicAppointment(c *gin.Context) {
	clinicID := c.Param("id")

	var req PublicAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen zorunlu alanları doldurunuz: " + err.Error()})
		return
	}

	// 1. Klinik ve Tedavi kontrolü
	var clinic models.Clinic
	if err := database.DB.First(&clinic, "id = ? AND is_active = ?", clinicID, true).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Klinik bulunamadı"})
		return
	}

	var treatment models.Treatment
	if err := database.DB.First(&treatment, "id = ? AND clinic_id = ?", req.TreatmentID, clinicID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz tedavi seçimi"})
		return
	}

	duration := treatment.DefaultDuration
	if duration <= 0 {
		duration = 30
	}
	endTime := req.StartTime.Add(time.Duration(duration) * time.Minute)

	// 2. Hekim ve Koltuk Ataması
	selectedDoctorID := req.DoctorID
	if selectedDoctorID == "" {
		// Boşta olan ilk hekimi bul
		var doctors []models.Doctor
		database.DB.Where("clinic_id = ?", clinicID).Find(&doctors)
		if len(doctors) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kliniğe kayıtlı hekim bulunmuyor"})
			return
		}

		for _, doc := range doctors {
			var overlapCount int64
			database.DB.Model(&models.Appointment{}).Where(
				"clinic_id = ? AND doctor_id = ? AND status != 'cancelled' AND ((start_time < ? AND end_time > ?))",
				clinicID, doc.ID, endTime, req.StartTime,
			).Count(&overlapCount)
			if overlapCount == 0 {
				selectedDoctorID = doc.ID
				break
			}
		}

		if selectedDoctorID == "" {
			selectedDoctorID = doctors[0].ID
		}
	}

	// Boş koltuk bul
	var chairs []models.Chair
	database.DB.Where("clinic_id = ? AND status != 'maintenance'", clinicID).Find(&chairs)
	selectedChairID := ""
	for _, ch := range chairs {
		var overlapCount int64
		database.DB.Model(&models.Appointment{}).Where(
			"clinic_id = ? AND chair_id = ? AND status != 'cancelled' AND ((start_time < ? AND end_time > ?))",
			clinicID, ch.ID, endTime, req.StartTime,
		).Count(&overlapCount)
		if overlapCount == 0 {
			selectedChairID = ch.ID
			break
		}
	}
	if selectedChairID == "" && len(chairs) > 0 {
		selectedChairID = chairs[0].ID
	}

	// 3. Hasta kaydını bul veya oluştur
	var patient models.Patient
	err := database.DB.Where("clinic_id = ? AND phone = ?", clinicID, req.Phone).First(&patient).Error
	if err != nil {
		patient = models.Patient{
			ClinicID:  clinicID,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Phone:     req.Phone,
			Email:     req.Email,
		}
		if err := database.DB.Create(&patient).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Hasta kaydı oluşturulamadı"})
			return
		}
	}

	// 4. Randevuyu oluştur
	appointment := models.Appointment{
		ClinicID:    clinicID,
		PatientID:   patient.ID,
		DoctorID:    selectedDoctorID,
		ChairID:     selectedChairID,
		TreatmentID: treatment.ID,
		StartTime:   req.StartTime,
		EndTime:     endTime,
		Notes:       req.Notes,
		Status:      "confirmed",
	}

	if err := database.DB.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Randevu oluşturulamadı: " + err.Error()})
		return
	}

	// Preload ilişkiler
	database.DB.Preload("Doctor.User").Preload("Treatment").Preload("Chair").First(&appointment, "id = ?", appointment.ID)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Randevunuz başarıyla oluşturuldu!",
		"appointment": appointment,
		"clinic":      clinic,
		"patient":     patient,
	})
}

