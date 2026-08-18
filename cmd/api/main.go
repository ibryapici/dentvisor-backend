package main

import (
	"log"
	"os"

	"dentvisor-backend/internal/handlers"
	"dentvisor-backend/internal/middleware"
	"dentvisor-backend/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Bilgi: .env dosyası bulunamadı, sistem çevre değişkenleri kullanılacak")
	}

	database.ConnectDB()
	database.SeedLocations()
	database.SeedSuperadmin()

	r := gin.Default()

	// CORS Ayarları (Basit haliyle)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Dentvisör API çalışıyor"})
	})

	authHandler := handlers.NewAuthHandler()

	// Public Routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.RegisterDoctor)
			auth.POST("/login", authHandler.Login)
		}

		locationHandler := handlers.NewLocationHandler()
		locations := api.Group("/locations")
		{
			locations.GET("/cities", locationHandler.GetCities)
			locations.GET("/cities/:id/districts", locationHandler.GetDistrictsByCity)
		}

		publicHandler := handlers.NewPublicHandler()
		public := api.Group("/public")
		{
			public.GET("/clinics", publicHandler.GetClinics)
			public.GET("/clinics/:id", publicHandler.GetClinicDetail)
			public.POST("/clinics/:id/claim", publicHandler.ClaimClinic)
		}
	}

	// Protected Routes (Sadece giriş yapmış kullanıcılar)
	protected := r.Group("/api/protected")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			role, _ := c.Get("role")
			c.JSON(200, gin.H{
				"message": "Korunan alana hoş geldiniz!",
				"user_id": userID,
				"role":    role,
			})
		})

		settingsHandler := handlers.NewSettingsHandler()
		settings := protected.Group("/settings")
		// Require admin or doctor for settings
		settings.Use(middleware.RequireRole("doctor", "admin"))
		{
			settings.GET("/clinic", settingsHandler.GetClinicProfile)
			settings.PUT("/clinic", settingsHandler.UpdateClinicProfile)

			settings.GET("/doctors", settingsHandler.GetDoctors)

			settings.GET("/treatments", settingsHandler.GetTreatments)
			settings.POST("/treatments", settingsHandler.AddTreatment)

			settings.GET("/chairs", settingsHandler.GetChairs)
			settings.POST("/chairs", settingsHandler.AddChair)
		}

		contentHandler := handlers.NewContentHandler()
		content := protected.Group("/content")
		content.Use(middleware.RequireRole("doctor", "admin", "secretary"))
		{
			content.GET("/reviews", contentHandler.GetReviews)
			content.PUT("/reviews/:id/status", contentHandler.UpdateReviewStatus)

			content.GET("/articles", contentHandler.GetArticles)
			content.POST("/articles", contentHandler.AddArticle)
		}

		patientHandler := handlers.NewPatientHandler()
		paymentHandler := handlers.NewPaymentHandler()
		treatmentPlanHandler := handlers.NewTreatmentPlanHandler()
		patients := protected.Group("/patients")
		patients.Use(middleware.RequireRole("doctor", "admin", "secretary"))
		{
			patients.GET("", patientHandler.GetPatients)
			patients.POST("", patientHandler.AddPatient)
			patients.GET("/:id", patientHandler.GetPatientDetail)
			patients.POST("/:id/dental-records", patientHandler.AddDentalRecord)
			
			// Payments
			patients.GET("/:id/payments", paymentHandler.GetPatientPayments)
			patients.POST("/:id/payments", paymentHandler.AddPayment)

			// Treatment Plans & Quotations
			patients.GET("/:id/treatment-plans", treatmentPlanHandler.GetPatientTreatmentPlans)
			patients.POST("/:id/treatment-plans", treatmentPlanHandler.CreateTreatmentPlan)
			patients.GET("/:id/treatment-plans/:planId", treatmentPlanHandler.GetTreatmentPlanDetail)
			patients.PUT("/:id/treatment-plans/:planId/status", treatmentPlanHandler.UpdateTreatmentPlanStatus)
			patients.DELETE("/:id/treatment-plans/:planId", treatmentPlanHandler.DeleteTreatmentPlan)
		}

		appointmentHandler := handlers.NewAppointmentHandler()
		appointments := protected.Group("/appointments")
		appointments.Use(middleware.RequireRole("doctor", "admin", "secretary"))
		{
			appointments.GET("", appointmentHandler.GetAppointments)
			appointments.POST("", appointmentHandler.AddAppointment)
			appointments.PUT("/:id/status", appointmentHandler.UpdateStatus)
		}

		labHandler := handlers.NewLabHandler()
		labs := protected.Group("")
		labs.Use(middleware.RequireRole("doctor", "admin", "secretary"))
		{
			// Laboratories
			labs.GET("/laboratories", labHandler.GetLaboratories)
			labs.POST("/laboratories", labHandler.CreateLaboratory)
			labs.PUT("/laboratories/:id", labHandler.UpdateLaboratory)
			labs.DELETE("/laboratories/:id", labHandler.DeleteLaboratory)

			// Lab Orders
			labs.GET("/lab-orders", labHandler.GetLabOrders)
			labs.POST("/lab-orders", labHandler.CreateLabOrder)
			labs.PUT("/lab-orders/:id/status", labHandler.UpdateLabOrderStatus)
			labs.PUT("/lab-orders/:id", labHandler.UpdateLabOrder)
			labs.DELETE("/lab-orders/:id", labHandler.DeleteLabOrder)

			// WhatsApp & AI Chatbot
			whatsappHandler := handlers.NewWhatsappHandler()
			labs.GET("/whatsapp/status", whatsappHandler.GetStatus)
			labs.POST("/whatsapp/connect", whatsappHandler.StartConnect)
			labs.POST("/whatsapp/confirm-scan", whatsappHandler.ConfirmScan)
			labs.POST("/whatsapp/disconnect", whatsappHandler.Disconnect)
			labs.PUT("/whatsapp/settings", whatsappHandler.UpdateSettings)
			labs.GET("/whatsapp/logs", whatsappHandler.GetMessageLogs)
			labs.POST("/whatsapp/send", whatsappHandler.SendMessage)
			labs.POST("/whatsapp/simulate-incoming", whatsappHandler.SimulateIncoming)
			labs.POST("/whatsapp/ai-generate-reminder", whatsappHandler.GenerateAiReminder)
		}

		// Sadece admin (doktor) görebilir (Klinik Yöneticisi)
		adminOnly := protected.Group("/admin")
		adminOnly.Use(middleware.RequireRole("doctor"))
		{
			reportHandler := handlers.NewReportHandler()
			adminOnly.GET("/reports/performance", reportHandler.GetPerformanceReport)
			adminOnly.GET("/reports/doctor-commissions", reportHandler.GetDoctorCommissions)
			adminOnly.PUT("/reports/doctors/:id/commission", reportHandler.UpdateDoctorCommission)
			adminOnly.POST("/reports/doctor-payouts", reportHandler.CreateDoctorPayout)
			adminOnly.GET("/reports/doctor-payouts", reportHandler.GetDoctorPayouts)

			automationHandler := handlers.NewAutomationHandler()
			adminOnly.POST("/automations/reminders", automationHandler.TriggerReminders)
		}

		// Sadece superadmin (Sistem Yöneticisi) görebilir
		systemOnly := protected.Group("/sistem")
		systemOnly.Use(middleware.RequireRole("superadmin"))
		{
			systemHandler := handlers.NewSystemHandler()
			systemOnly.GET("/dashboard", systemHandler.GetDashboardStats)
			systemOnly.GET("/clinics", systemHandler.GetClinics)
			systemOnly.PUT("/clinics/:id", systemHandler.UpdateClinic)
			systemOnly.PUT("/clinics/:id/status", systemHandler.UpdateClinicStatus)
			systemOnly.POST("/impersonate", systemHandler.Impersonate)
			systemOnly.GET("/patients", systemHandler.GetAllPatients)
			systemOnly.GET("/appointments", systemHandler.GetAllAppointments)

			// Location Management
			systemOnly.GET("/locations/cities", systemHandler.GetSystemCities)
			systemOnly.POST("/locations/cities", systemHandler.CreateCity)
			systemOnly.PUT("/locations/cities/:id", systemHandler.UpdateCity)
			systemOnly.DELETE("/locations/cities/:id", systemHandler.DeleteCity)

			systemOnly.GET("/locations/cities/:cityId/districts", systemHandler.GetSystemDistricts)
			systemOnly.POST("/locations/cities/:cityId/districts", systemHandler.CreateDistrict)
			systemOnly.PUT("/locations/districts/:id", systemHandler.UpdateDistrict)
			systemOnly.DELETE("/locations/districts/:id", systemHandler.DeleteDistrict)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Sunucu %s portunda başlatılıyor...", port)
	r.Run(":" + port)
}
