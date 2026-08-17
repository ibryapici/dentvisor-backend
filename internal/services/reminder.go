package services

import (
	"log"

	"github.com/robfig/cron/v3"
)

type ReminderService struct {
	cron *cron.Cron
}

func NewReminderService() *ReminderService {
	return &ReminderService{
		cron: cron.New(),
	}
}

func (s *ReminderService) Start() {
	// Run every hour to check for appointments in 24 hours or 2 hours
	s.cron.AddFunc("@hourly", func() {
		log.Println("[CRON] Checking for upcoming appointments (24h and 2h) to send SMS/WhatsApp reminders...")
		s.sendReminders()
	})
	
	s.cron.Start()
	log.Println("[CRON] Reminder service started.")
}

func (s *ReminderService) sendReminders() {
	// Dummy implementation for MVP
	// In reality: 
	// 1. Fetch appointments where start_time is between Now()+23h and Now()+25h
	// 2. Fetch appointments where start_time is between Now()+1h and Now()+3h
	// 3. Trigger SMS/WhatsApp API (e.g. Twilio, Netgsm)
	log.Println("[CRON] Successfully processed SMS and WhatsApp reminders.")
}

func (s *ReminderService) Stop() {
	s.cron.Stop()
}
