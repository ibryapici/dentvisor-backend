package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type WhatsappHandler struct{}

func NewWhatsappHandler() *WhatsappHandler {
	return &WhatsappHandler{}
}

// GET /api/protected/whatsapp/status
func (h *WhatsappHandler) GetStatus(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var setting models.WhatsappSetting
	err := database.DB.First(&setting, "clinic_id = ?", clinicID).Error
	if err != nil {
		// Varsayılan ayar oluşturalım
		setting = models.WhatsappSetting{
			ClinicID:            clinicID,
			Status:              "disconnected",
			IsAiEnabled:         true,
			AiBotName:           "Dentvisör AI Klinik Asistanı",
			AiInstructions:      "Kliniğimiz haftanın 6 günü 09:00 - 19:00 saatleri arasında açıktır. Estetik gülüş tasarımı, implant, zirkonyum kaplama ve kanal tedavisi konularında uzmanız. Randevu taleplerini nazikçe al ve hasta bilgilerini (ad, soyad, telefon) iste.",
			AutoReminderEnabled: true,
			ReminderHoursBefore: 24,
			AiReminderTone:      "warm_professional",
			PostCareEnabled:     true,
			GoogleReviewLink:    "https://g.page/r/dentvisor/review",
		}
		database.DB.Create(&setting)
	}

	c.JSON(http.StatusOK, setting)
}

// POST /api/protected/whatsapp/connect
func (h *WhatsappHandler) StartConnect(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var setting models.WhatsappSetting
	database.DB.First(&setting, "clinic_id = ?", clinicID)

	// Simüle edilmiş canlı WhatsApp Web QR Kodu
	timestamp := time.Now().Unix()
	qrCodePayload := fmt.Sprintf("2@%d,DENTVISOR_WA_%s,SECURE_HANDSHAKE", timestamp, clinicID[:8])

	setting.Status = "qr_ready"
	setting.QRCode = qrCodePayload
	database.DB.Save(&setting)

	c.JSON(http.StatusOK, gin.H{
		"status":  "qr_ready",
		"qr_code": qrCodePayload,
		"message": "QR kod oluşturuldu. Lütfen WhatsApp uygulamasından Bağlı Cihazlar > Cihaz Bağla adımı ile okutunuz.",
	})
}

// POST /api/protected/whatsapp/confirm-scan (QR okutulunca oturumu aktif etme simülasyonu)
func (h *WhatsappHandler) ConfirmScan(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	c.ShouldBindJSON(&req)
	if req.PhoneNumber == "" {
		req.PhoneNumber = "+90 532 555 0123"
	}

	var setting models.WhatsappSetting
	database.DB.First(&setting, "clinic_id = ?", clinicID)
	setting.Status = "connected"
	setting.PhoneNumber = req.PhoneNumber
	setting.QRCode = ""
	database.DB.Save(&setting)

	c.JSON(http.StatusOK, gin.H{
		"status":       "connected",
		"phone_number": setting.PhoneNumber,
		"message":      "WhatsApp Business hesabınız başarıyla bağlandı!",
	})
}

// POST /api/protected/whatsapp/disconnect
func (h *WhatsappHandler) Disconnect(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var setting models.WhatsappSetting
	database.DB.First(&setting, "clinic_id = ?", clinicID)
	setting.Status = "disconnected"
	setting.QRCode = ""
	setting.PhoneNumber = ""
	database.DB.Save(&setting)

	c.JSON(http.StatusOK, gin.H{"message": "WhatsApp bağlantısı başarıyla kesildi."})
}

// PUT /api/protected/whatsapp/settings
func (h *WhatsappHandler) UpdateSettings(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var req models.WhatsappSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	var setting models.WhatsappSetting
	database.DB.First(&setting, "clinic_id = ?", clinicID)

	setting.IsAiEnabled = req.IsAiEnabled
	setting.AiBotName = req.AiBotName
	setting.AiInstructions = req.AiInstructions
	setting.AutoReminderEnabled = req.AutoReminderEnabled
	setting.ReminderHoursBefore = req.ReminderHoursBefore
	setting.AiReminderTone = req.AiReminderTone
	setting.PostCareEnabled = req.PostCareEnabled
	setting.GoogleReviewLink = req.GoogleReviewLink

	database.DB.Save(&setting)
	c.JSON(http.StatusOK, setting)
}

// GET /api/protected/whatsapp/logs
func (h *WhatsappHandler) GetMessageLogs(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var logs []models.WhatsappMessageLog
	database.DB.Where("clinic_id = ?", clinicID).
		Order("created_at desc").
		Limit(50).
		Find(&logs)

	c.JSON(http.StatusOK, logs)
}

// POST /api/protected/whatsapp/send
func (h *WhatsappHandler) SendMessage(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon ve mesaj alanları zorunludur"})
		return
	}

	log := models.WhatsappMessageLog{
		ClinicID:      clinicID,
		Direction:     "outgoing",
		Phone:         req.Phone,
		SenderName:    "Klinik Yetkilisi",
		Message:       req.Message,
		IsAiGenerated: false,
		Status:        "sent",
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&log)

	c.JSON(http.StatusOK, gin.H{"message": "Mesaj gönderildi", "log": log})
}

// POST /api/protected/whatsapp/simulate-incoming (AI Chatbot Test Simülatörü)
func (h *WhatsappHandler) SimulateIncoming(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var req struct {
		PatientName string `json:"patient_name"`
		Phone       string `json:"phone"`
		Message     string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mesaj alanı zorunludur"})
		return
	}

	if req.PatientName == "" {
		req.PatientName = "Hasta (Ziyaretçi)"
	}
	if req.Phone == "" {
		req.Phone = "+90 544 123 4567"
	}

	// 1. Gelen mesajı kaydet
	incomingLog := models.WhatsappMessageLog{
		ClinicID:      clinicID,
		Direction:     "incoming",
		Phone:         req.Phone,
		SenderName:    req.PatientName,
		Message:       req.Message,
		IsAiGenerated: false,
		Status:        "received",
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&incomingLog)

	// 2. Klinik ve AI ayarlarını al
	var setting models.WhatsappSetting
	database.DB.First(&setting, "clinic_id = ?", clinicID)

	var clinic models.Clinic
	database.DB.First(&clinic, "id = ?", clinicID)

	// 3. AI Asistanı Yanıt Üretimi
	var aiReply string
	userMsgLower := strings.ToLower(req.Message)

	if strings.Contains(userMsgLower, "randevu") || strings.Contains(userMsgLower, "saat") || strings.Contains(userMsgLower, "müsait") {
		aiReply = fmt.Sprintf("Merhaba %s! %s olarak size yardımcı olmaktan memnuniyet duyarız. 🦷 Kliniğimiz %s saatleri arasında hizmet vermektedir. Size en uygun gün ve saati belirtirseniz hekimlerimizin müsaitlik durumunu kontrol edip randevunuzu oluşturalım.", req.PatientName, clinic.Name, clinic.WorkingHours)
	} else if strings.Contains(userMsgLower, "fiyat") || strings.Contains(userMsgLower, "ücret") || strings.Contains(userMsgLower, "kaç para") || strings.Contains(userMsgLower, "implant") || strings.Contains(userMsgLower, "zirkonyum") {
		aiReply = fmt.Sprintf("Merhaba! %s bünyesinde uygulanan tedavilerde kişiye özel ağız ve diş yapısına göre en uygun tedavi planı belirlenmektedir. Kesin maliyet için hekimimizin ücretsiz ön muayenesine davet etmek isteriz. Randevu planlayalım mı?", clinic.Name)
	} else if strings.Contains(userMsgLower, "adres") || strings.Contains(userMsgLower, "nerede") || strings.Contains(userMsgLower, "konum") {
		aiReply = fmt.Sprintf("Kliniğimizin açık adresi: %s. 📍 Kolayca ulaşım sağlayabilirsiniz. Gelmeden önce randevunuzu teyit etmenizi rica ederiz.", clinic.Address)
	} else if strings.Contains(userMsgLower, "merhaba") || strings.Contains(userMsgLower, "selam") || strings.Contains(userMsgLower, "iyi günler") {
		aiReply = fmt.Sprintf("Merhaba %s, %s kliniğine hoş geldiniz! ✨ Ben kliniğinizin yapay zeka asistanıyım. Tedavilerimiz, randevu veya muayene ile ilgili size nasıl yardımcı olabilirim?", req.PatientName, clinic.Name)
	} else {
		aiReply = fmt.Sprintf("Mesajınız için teşekkür ederiz. %s uzman hekimlerimiz ve ekibimiz mesajınızı inceleyip en kısa sürede size dönüş yapacaktır. Acil durumlarda bizi doğrudan %s numarasından arayabilirsiniz. Sağlıklı günler dileriz! 🌿", clinic.Name, clinic.Phone)
	}

	// 4. AI Yanıtını Logla
	outgoingLog := models.WhatsappMessageLog{
		ClinicID:      clinicID,
		Direction:     "outgoing",
		Phone:         req.Phone,
		SenderName:    setting.AiBotName,
		Message:       aiReply,
		IsAiGenerated: true,
		Status:        "sent",
		CreatedAt:     time.Now().Add(time.Second * 1),
	}
	database.DB.Create(&outgoingLog)

	c.JSON(http.StatusOK, gin.H{
		"incoming": incomingLog,
		"reply":    outgoingLog,
	})
}

// POST /api/protected/whatsapp/ai-generate-reminder (AI Kişiselleştirilmiş Hatırlatma)
func (h *WhatsappHandler) GenerateAiReminder(c *gin.Context) {
	clinicID, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik yetkisi bulunamadı"})
		return
	}

	var req struct {
		PatientName string `json:"patient_name" binding:"required"`
		DoctorName  string `json:"doctor_name"`
		Date        string `json:"date" binding:"required"`
		Time        string `json:"time" binding:"required"`
		Treatment   string `json:"treatment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Eksik parametreler"})
		return
	}

	var clinic models.Clinic
	database.DB.First(&clinic, "id = ?", clinicID)

	treatmentInfo := ""
	if req.Treatment != "" {
		treatmentInfo = fmt.Sprintf("%s tedaviniz için ", req.Treatment)
	}

	doctorInfo := ""
	if req.DoctorName != "" {
		doctorInfo = fmt.Sprintf("Dt. %s ile ", req.DoctorName)
	}

	reminderMessage := fmt.Sprintf("🦷 *%s - Randevu Hatırlatması*\n\nSayın *%s*,\n\n%s tarihinde saat *%s* için %s%srandevunuz bulunmaktadır.\n\n✨ Tedavinize zamanında başlayabilmemiz için randevu saatinizden 10 dakika önce kliniğimizde olmanızı rica ederiz.\n\n📍 *Adres:* %s\n📞 *İletişim:* %s\n\nRandevunuza katılım durumunuzu bu mesaja *'EVET'* veya ertelemek isterseniz *'ERTELE'* yazarak bize iletebilirsiniz. Sağlıklı gülüşler dileriz! 🌿",
		clinic.Name, req.PatientName, req.Date, req.Time, doctorInfo, treatmentInfo, clinic.Address, clinic.Phone)

	c.JSON(http.StatusOK, gin.H{
		"generated_message": reminderMessage,
		"tone":              "warm_professional",
	})
}
