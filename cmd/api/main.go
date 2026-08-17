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
		patients := protected.Group("/patients")
		patients.Use(middleware.RequireRole("doctor", "admin", "secretary"))
		{
			patients.GET("", patientHandler.GetPatients)
			patients.POST("", patientHandler.AddPatient)
			patients.GET("/:id", patientHandler.GetPatientDetail)
			patients.POST("/:id/dental-records", patientHandler.AddDentalRecord)
		}

		appointmentHandler := handlers.NewAppointmentHandler()
		appointments := protected.Group("/appointments")
		appointments.Use(middleware.RequireRole("doctor", "admin", "secretary"))
		{
			appointments.GET("", appointmentHandler.GetAppointments)
			appointments.POST("", appointmentHandler.AddAppointment)
			appointments.PUT("/:id/status", appointmentHandler.UpdateStatus)
		}

		// Sadece admin (doktor) görebilir
		adminOnly := protected.Group("/admin")
		adminOnly.Use(middleware.RequireRole("doctor"))
		{
			reportHandler := handlers.NewReportHandler()
			adminOnly.GET("/reports/performance", reportHandler.GetPerformanceReport)

			automationHandler := handlers.NewAutomationHandler()
			adminOnly.POST("/automations/reminders", automationHandler.TriggerReminders)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Sunucu %s portunda başlatılıyor...", port)
	r.Run(":" + port)
}
