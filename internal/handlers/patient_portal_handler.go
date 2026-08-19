package handlers

import (
	"net/http"
	"time"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type PatientPortalHandler struct{}

func NewPatientPortalHandler() *PatientPortalHandler {
	return &PatientPortalHandler{}
}

// GetPatientDashboard returns full aggregated multi-clinic data for the logged-in patient
func (h *PatientPortalHandler) GetPatientDashboard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz erişim"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// 1. Find all Patient profiles belonging to this User (across all clinics)
	var patientProfiles []models.Patient
	database.DB.Preload("Clinic").
		Where("user_id = ? OR email = ? OR (phone != '' AND phone = ?)", user.ID, user.Email, user.Phone).
		Find(&patientProfiles)

	if len(patientProfiles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"user":            user,
			"clinics":         []models.Clinic{},
			"dental_records":  []gin.H{},
			"appointments":    []models.Appointment{},
			"treatment_plans": []models.TreatmentPlan{},
			"balances":        []gin.H{},
			"total_balance":   0,
			"consents":        []models.PatientConsent{},
		})
		return
	}

	// 2. Collect Patient IDs and unique Clinics
	patientIDs := make([]string, 0, len(patientProfiles))
	clinicMap := make(map[string]models.Clinic)
	clinicBalances := make([]gin.H, 0)
	totalBalance := 0.0

	for _, p := range patientProfiles {
		patientIDs = append(patientIDs, p.ID)
		if p.Clinic != nil {
			clinicMap[p.Clinic.ID] = *p.Clinic
			clinicBalances = append(clinicBalances, gin.H{
				"clinic_id":   p.Clinic.ID,
				"clinic_name": p.Clinic.Name,
				"balance":     p.Balance,
			})
			totalBalance += p.Balance
		}
	}

	clinics := make([]models.Clinic, 0, len(clinicMap))
	for _, cl := range clinicMap {
		clinics = append(clinics, cl)
	}

	// 3. Fetch all DentalRecords across all patient profiles
	var rawRecords []models.DentalRecord
	database.DB.Preload("Treatment").
		Preload("Doctor.User").
		Where("patient_id IN ?", patientIDs).
		Order("date desc").
		Find(&rawRecords)

	// Enrich dental records with clinic details
	dentalRecords := make([]gin.H, 0, len(rawRecords))
	for _, rec := range rawRecords {
		var pClinicName, pClinicID string
		for _, p := range patientProfiles {
			if p.ID == rec.PatientID && p.Clinic != nil {
				pClinicName = p.Clinic.Name
				pClinicID = p.Clinic.ID
				break
			}
		}

		docName := "Hekim"
		if rec.Doctor.User.FirstName != "" {
			docName = "Dt. " + rec.Doctor.User.FirstName + " " + rec.Doctor.User.LastName
		}

		treatmentName := "Tedavi"
		if rec.Treatment.Name != "" {
			treatmentName = rec.Treatment.Name
		}

		dentalRecords = append(dentalRecords, gin.H{
			"id":             rec.ID,
			"patient_id":     rec.PatientID,
			"clinic_id":      pClinicID,
			"clinic_name":    pClinicName,
			"doctor_name":    docName,
			"treatment_name": treatmentName,
			"tooth_number":   rec.ToothNumber,
			"status":         rec.Status,
			"notes":          rec.Notes,
			"date":           rec.Date,
		})
	}

	// 4. Fetch all Appointments across all patient profiles
	var appointments []models.Appointment
	database.DB.Preload("Clinic").
		Preload("Doctor.User").
		Preload("Treatment").
		Where("patient_id IN ?", patientIDs).
		Order("start_time desc").
		Find(&appointments)

	// 5. Fetch all Treatment Plans across all patient profiles
	var plans []models.TreatmentPlan
	database.DB.Preload("Items").
		Preload("Clinic").
		Preload("Doctor.User").
		Where("patient_id IN ?", patientIDs).
		Order("created_at desc").
		Find(&plans)

	// 6. Fetch all Consents
	var consents []models.PatientConsent
	database.DB.Preload("TargetClinic").
		Where("patient_id IN ?", patientIDs).
		Order("created_at desc").
		Find(&consents)

	c.JSON(http.StatusOK, gin.H{
		"user":            user,
		"patient":         patientProfiles[0],
		"profiles":        patientProfiles,
		"clinics":         clinics,
		"dental_records":  dentalRecords,
		"appointments":    appointments,
		"treatment_plans": plans,
		"balances":        clinicBalances,
		"total_balance":   totalBalance,
		"consents":        consents,
	})
}

// RequestHistoryConsent allows a clinic doctor to send a consent request to the patient
func (h *PatientPortalHandler) RequestHistoryConsent(c *gin.Context) {
	patientID := c.Param("id")
	clinicID, _ := c.Get("clinic_id")

	if clinicID == nil || clinicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Klinik bilgisi bulunamadı"})
		return
	}

	// Check if already approved or pending
	var existing models.PatientConsent
	err := database.DB.Where("patient_id = ? AND target_clinic_id = ? AND status = 'approved' AND expires_at > ?",
		patientID, clinicID, time.Now()).First(&existing).Error

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hastanın geçmiş izni zaten aktif",
			"consent": existing,
		})
		return
	}

	consent := models.PatientConsent{
		PatientID:      patientID,
		TargetClinicID: clinicID.(string),
		Status:         "pending",
		ExpiresAt:      time.Now().AddDate(0, 1, 0), // 30 days valid
	}

	database.DB.Create(&consent)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Hastaya WhatsApp / SMS üzerinden geçmiş diş şeması onay isteği gönderildi.",
		"consent": consent,
	})
}

// RespondConsent allows the patient to approve or reject a consent request
func (h *PatientPortalHandler) RespondConsent(c *gin.Context) {
	consentID := c.Param("id")

	var req struct {
		Action string `json:"action" binding:"required"` // 'approve' or 'revoke'
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}

	var consent models.PatientConsent
	if err := database.DB.First(&consent, "id = ?", consentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "İzin talebi bulunamadı"})
		return
	}

	if req.Action == "approve" {
		consent.Status = "approved"
		consent.ExpiresAt = time.Now().AddDate(0, 1, 0)
	} else {
		consent.Status = "revoked"
	}

	database.DB.Save(&consent)

	c.JSON(http.StatusOK, gin.H{
		"message": "İzin durumu başarıyla güncellendi",
		"consent": consent,
	})
}

// GetConsultationDentalHistory allows a doctor in a clinic with approved consent to view the patient's full medical records
func (h *PatientPortalHandler) GetConsultationDentalHistory(c *gin.Context) {
	patientID := c.Param("id")
	clinicID, _ := c.Get("clinic_id")

	var targetPatient models.Patient
	if err := database.DB.First(&targetPatient, "id = ?", patientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hasta bulunamadı"})
		return
	}

	// 1. Check if same clinic (own clinic has natural access)
	isOwnClinic := targetPatient.ClinicID == clinicID

	if !isOwnClinic {
		// Verify approved consent
		var consent models.PatientConsent
		err := database.DB.Where("patient_id = ? AND target_clinic_id = ? AND status = 'approved' AND expires_at > ?",
			patientID, clinicID, time.Now()).First(&consent).Error

		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":           "Bu hastanın geçmiş klinik kayıtlarını görüntülemek için açık rıza (onay) izni gereklidir.",
				"consent_required": true,
			})
			return
		}
	}

	// Find all patient records matching email or phone
	var allProfiles []models.Patient
	database.DB.Preload("Clinic").
		Where("email = ? OR (phone != '' AND phone = ?)", targetPatient.Email, targetPatient.Phone).
		Find(&allProfiles)

	profileIDs := make([]string, 0, len(allProfiles))
	for _, p := range allProfiles {
		profileIDs = append(profileIDs, p.ID)
	}

	// Fetch medical records (strictly excluding financial prices)
	var records []models.DentalRecord
	database.DB.Preload("Treatment").
		Preload("Doctor.User").
		Where("patient_id IN ?", profileIDs).
		Order("date desc").
		Find(&records)

	history := make([]gin.H, 0, len(records))
	for _, r := range records {
		var clName string
		for _, p := range allProfiles {
			if p.ID == r.PatientID && p.Clinic != nil {
				clName = p.Clinic.Name
				break
			}
		}

		docName := "Hekim"
		if r.Doctor.User.FirstName != "" {
			docName = "Dt. " + r.Doctor.User.FirstName + " " + r.Doctor.User.LastName
		}

		history = append(history, gin.H{
			"id":             r.ID,
			"tooth_number":   r.ToothNumber,
			"treatment_name": r.Treatment.Name,
			"clinic_name":    clName,
			"doctor_name":    docName,
			"status":         r.Status,
			"notes":          r.Notes,
			"date":           r.Date,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"is_own_clinic":      isOwnClinic,
		"consultation_records": history,
	})
}

// CreatePatientAppointment creates an appointment booking from the patient portal
func (h *PatientPortalHandler) CreatePatientAppointment(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		ClinicID    string    `json:"clinic_id" binding:"required"`
		DoctorID    string    `json:"doctor_id"`
		TreatmentID string    `json:"treatment_id"`
		StartTime   time.Time `json:"start_time" binding:"required"`
		Notes       string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz randevu formu"})
		return
	}

	// Find patient profile for this clinic
	var patient models.Patient
	err := database.DB.Where("user_id = ? AND clinic_id = ?", userID, req.ClinicID).First(&patient).Error
	if err != nil {
		// Fallback by user email
		var u models.User
		database.DB.First(&u, "id = ?", userID)
		err = database.DB.Where("email = ? AND clinic_id = ?", u.Email, req.ClinicID).First(&patient).Error
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu klinikte kayıtlı hasta dosyanız bulunamadı"})
		return
	}

	endTime := req.StartTime.Add(30 * time.Minute)
	appointment := models.Appointment{
		ClinicID:    req.ClinicID,
		PatientID:   patient.ID,
		DoctorID:    req.DoctorID,
		StartTime:   req.StartTime,
		EndTime:     endTime,
		Status:      "pending",
		Notes:       req.Notes,
	}

	if err := database.DB.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Randevu oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Randevu talebiniz kliniğe iletildi.",
		"appointment": appointment,
	})
}

// AcceptTreatmentPlan updates a treatment plan status to 'accepted' when confirmed by patient
func (h *PatientPortalHandler) AcceptTreatmentPlan(c *gin.Context) {
	planID := c.Param("planId")

	var plan models.TreatmentPlan
	if err := database.DB.First(&plan, "id = ?", planID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tedavi planı bulunamadı"})
		return
	}

	plan.Status = "accepted"
	database.DB.Save(&plan)

	c.JSON(http.StatusOK, gin.H{
		"message": "Tedavi planını ve fiyat teklifini başarıyla onayladınız.",
		"plan":    plan,
	})
}

