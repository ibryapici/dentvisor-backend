package services

import (
	"log"
	"time"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// SendSMS simüle edilmiş (mock) bir SMS gönderme metodudur.
func (s *NotificationService) SendSMS(phone, message string) error {
	// İleride buraya Twilio, Netgsm, vb. entegrasyonu gelecek.
	log.Printf("[SMS MOCK] %s numarasına SMS gönderiliyor...", phone)
	time.Sleep(500 * time.Millisecond) // Simüle edilmiş gecikme
	log.Printf("[SMS MOCK] Başarılı: %s", message)
	return nil
}

// SendWhatsApp simüle edilmiş (mock) bir WhatsApp mesajı gönderme metodudur.
func (s *NotificationService) SendWhatsApp(phone, message string) error {
	// İleride buraya WhatsApp Business API veya 3. parti (Örn: Twilio, WATI) entegrasyonu gelecek.
	log.Printf("[WhatsApp MOCK] %s numarasına WhatsApp mesajı gönderiliyor...", phone)
	time.Sleep(500 * time.Millisecond) // Simüle edilmiş gecikme
	log.Printf("[WhatsApp MOCK] Başarılı: %s", message)
	return nil
}
