package handlers

import (
	"net/http"

	"dentvisor-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AutomationHandler struct{}

func NewAutomationHandler() *AutomationHandler {
	return &AutomationHandler{}
}

// POST /api/protected/admin/automations/reminders
func (h *AutomationHandler) TriggerReminders(c *gin.Context) {
	_, ok := getClinicID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Klinik ID bulunamadı"})
		return
	}

	// MOCK: Burada normalde veritabanından yarınki randevular çekilip Twilio veya WhatsApp API'sine (örn: Meta Graph API) istek atılır.
	// Şimdilik işlemi simüle ediyoruz.
	notificationService := services.NewNotificationService()
	_ = notificationService.SendSMS("+905550000000", "Yarın saat 14:00'te Dentvisor Kliniği'nde randevunuz bulunmaktadır. Lütfen geç kalmayınız.")
	_ = notificationService.SendWhatsApp("+905550000001", "Merhaba! Yarın saat 15:30'daki diş hekimi randevunuzu hatırlatmak isteriz. İyi günler!")

	c.JSON(http.StatusOK, gin.H{
		"message": "Otomatik hatırlatıcılar tetiklendi",
		"details": gin.H{
			"sms_sent":      5,
			"whatsapp_sent": 3,
			"failed":        0,
		},
	})
}
