package database

import (
	"fmt"
	"log"
	"os"

	"dentvisor-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Istanbul",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("UYARI: Veritabanına bağlanılamadı. DB ayarlarınızı kontrol edin.", err)
		return
	}

	DB = db
	log.Println("PostgreSQL veritabanına başarıyla bağlanıldı!")

	// Enable UUID extension
	if err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Println("UUID eklentisi oluşturulamadı:", err)
	}

	// GORM AutoMigrate: Modelleri veritabanı tablolarıyla senkronize eder
	err = DB.AutoMigrate(
		&models.City{},
		&models.District{},
		&models.Clinic{},
		&models.User{},
		&models.Doctor{},
		&models.Patient{},
		&models.Treatment{},
		&models.Chair{},
		&models.Review{},
		&models.Article{},
		&models.Appointment{},
		&models.Payment{},
		&models.DentalRecord{},
		&models.TreatmentPlan{},
		&models.TreatmentPlanItem{},
		&models.Laboratory{},
		&models.LabOrder{},
		&models.WhatsappSetting{},
		&models.WhatsappMessageLog{},
		&models.DoctorPayout{},
	)
	if err != nil {
		log.Println("AutoMigrate hatası:", err)
	} else {
		log.Println("AutoMigrate başarıyla tamamlandı (Yeni modeller eklendi).")
	}
}
