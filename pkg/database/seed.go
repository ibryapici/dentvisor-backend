package database

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"dentvisor-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// PHPMyAdmin Export Format structs
type PMAExport struct {
	Type string            `json:"type"`
	Name string            `json:"name"`
	Data []json.RawMessage `json:"data"`
}

type İlData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type İlçeData struct {
	ID    string `json:"id"`
	IlID  string `json:"il_id"`
	Name  string `json:"name"`
}

func SeedLocations() {
	var count int64
	DB.Model(&models.City{}).Count(&count)
	if count >= 81 {
		return // Data already seeded completely
	}

	log.Println("Gerçek İl ve İlçe verileri okunuyor (Yerel dosyalardan)...")

	// Sadece lokasyon tablolarını temizle (CASCADE kullanma - users/clinics silinmesin!)
	// FK kısıtını geçici devre dışı bırakarak sıralı sil
	if err := DB.Exec("DELETE FROM districts;").Error; err != nil {
		log.Println("Districts temizlenirken hata:", err)
		return
	}
	if err := DB.Exec("DELETE FROM cities;").Error; err != nil {
		log.Println("Cities temizlenirken hata:", err)
		return
	}

	// 1. İlleri Oku
	ilBytes, err := os.ReadFile("db/seed/data/il.json")
	if err != nil {
		log.Println("İl verisi okunamadı:", err)
		return
	}

	var pmaIl []PMAExport
	if err := json.Unmarshal(ilBytes, &pmaIl); err != nil {
		log.Println("İl JSON parse hatası:", err)
		return
	}

	// 2. İlçeleri Oku
	ilceBytes, err := os.ReadFile("db/seed/data/ilce.json")
	if err != nil {
		log.Println("İlçe verisi okunamadı:", err)
		return
	}

	var pmaIlce []PMAExport
	if err := json.Unmarshal(ilceBytes, &pmaIlce); err != nil {
		log.Println("İlçe JSON parse hatası:", err)
		return
	}

	// Verileri veritabanına ekle
	for _, export := range pmaIl {
		if export.Type == "table" && export.Name == "il" {
			for _, rawItem := range export.Data {
				var il İlData
				json.Unmarshal(rawItem, &il)
				id, _ := strconv.Atoi(il.ID)
				
				city := models.City{
					ID:   uint(id),
					Name: il.Name,
				}
				DB.Create(&city)
			}
		}
	}

	for _, export := range pmaIlce {
		if export.Type == "table" && export.Name == "ilce" {
			// Batch insert için slice
			var districts []models.District
			for _, rawItem := range export.Data {
				var ilce İlçeData
				json.Unmarshal(rawItem, &ilce)
				
				id, _ := strconv.Atoi(ilce.ID)
				ilId, _ := strconv.Atoi(ilce.IlID)
				
				districts = append(districts, models.District{
					ID:     uint(id),
					CityID: uint(ilId),
					Name:   ilce.Name,
				})
			}
			// Toplu kayıt
			DB.CreateInBatches(districts, 100)
		}
	}

	log.Println("Tüm 81 il ve ilçeleri veritabanına başarıyla eklendi!")
}

func SeedSuperadmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "superadmin").Count(&count)
	if count > 0 {
		return // Superadmin already exists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("superadmin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Superadmin şifresi oluşturulamadı:", err)
		return
	}

	superadmin := models.User{
		Email:        "superadmin@dentvisor.com",
		PasswordHash: string(hashedPassword),
		Role:         "superadmin",
		FirstName:    "Sistem",
		LastName:     "Yöneticisi",
	}

	if err := DB.Create(&superadmin).Error; err != nil {
		log.Println("Superadmin hesabı oluşturulamadı:", err)
		return
	}

	log.Println("Varsayılan superadmin hesabı oluşturuldu (superadmin@dentvisor.com / superadmin123)")
}
