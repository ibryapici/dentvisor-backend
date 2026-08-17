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
		api.POST("/auth/register", authHandler.RegisterDoctor)
		api.POST("/auth/login", authHandler.Login)
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

		// Sadece admin (doktor) görebilir
		adminOnly := protected.Group("/admin")
		adminOnly.Use(middleware.RequireRole("doctor"))
		{
			adminOnly.GET("/reports", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Finansal raporlar - Sadece Doktor/Admin"})
			})
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Sunucu %s portunda başlatılıyor...", port)
	r.Run(":" + port)
}
