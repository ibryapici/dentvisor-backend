package database

import (
	"log"
	"time"

	"dentvisor-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedDemoData() {
	// 1. admin@test.com kullanıcısını ve kliniğini bul veya oluştur
	var user models.User
	err := DB.Where("email = ?", "admin@test.com").First(&user).Error

	var clinic models.Clinic
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	cityID := uint(34)     // İstanbul
	districtID := uint(1663) // Şişli

	if err != nil {
		// Yeni klinik oluştur
		clinic = models.Clinic{
			Name:         "Özel Dentvisör Ağız ve Diş Sağlığı Polikliniği",
			Phone:        "0212 555 40 40",
			Address:      "Valikonağı Cad. No:42/B, Nişantaşı, Şişli / İstanbul",
			AboutText:    "Dentvisör Polikliniği olarak 15 yılı aşkın süredir estetik diş hekimliği, implantoloji, ortodonti ve çocuk diş hekimliği alanlarında multidisipliner uzman kadromuz ile hizmet vermekteyiz.",
			WorkingHours: "Pzt - Cmt: 09:00 - 20:00",
			IsActive:     true,
			CityID:       &cityID,
			DistrictID:   &districtID,
		}
		if err := DB.Create(&clinic).Error; err != nil {
			log.Println("Demo klinik oluşturulamadı:", err)
			return
		}

		user = models.User{
			ClinicID:     &clinic.ID,
			Email:        "admin@test.com",
			PasswordHash: string(hashedPassword),
			Role:         "doctor",
			FirstName:    "Ahmet",
			LastName:     "Yılmaz",
			Phone:        "0532 555 4040",
		}
		DB.Create(&user)
	} else {
		// Var olan kullanıcının şifresini password123 olarak güncelle ve kliniğini bağla
		user.PasswordHash = string(hashedPassword)
		if user.ClinicID != nil {
			DB.First(&clinic, "id = ?", *user.ClinicID)
		} else {
			clinic = models.Clinic{
				Name:         "Özel Dentvisör Ağız ve Diş Sağlığı Polikliniği",
				Phone:        "0212 555 40 40",
				Address:      "Valikonağı Cad. No:42/B, Nişantaşı, Şişli / İstanbul",
				AboutText:    "Dentvisör Polikliniği olarak 15 yılı aşkın süredir estetik diş hekimliği, implantoloji ve ortodonti alanlarında uzman kadromuz ile hizmet vermekteyiz.",
				WorkingHours: "Pzt - Cmt: 09:00 - 20:00",
				IsActive:     true,
				CityID:       &cityID,
				DistrictID:   &districtID,
			}
			DB.Create(&clinic)
			user.ClinicID = &clinic.ID
		}
		DB.Save(&user)
	}

	clinicID := clinic.ID

	// 2. Doktorlar (3 Uzman Hekim)
	var docCount int64
	DB.Model(&models.Doctor{}).Where("clinic_id = ?", clinicID).Count(&docCount)
	if docCount < 3 {
		// Ana doktor (Dr. Ahmet Yılmaz)
		var doc1 models.Doctor
		if err := DB.Where("user_id = ?", user.ID).First(&doc1).Error; err != nil {
			doc1 = models.Doctor{
				UserID:         user.ID,
				ClinicID:       clinicID,
				Title:          "Uzm. Dt.",
				Specialty:      "Ortodonti & Çene Cerrahisi",
				CommissionRate: 35.0,
				BaseSalary:     40000.0,
			}
			DB.Create(&doc1)
		} else {
			doc1.CommissionRate = 35.0
			doc1.BaseSalary = 40000.0
			doc1.Title = "Uzm. Dt."
			doc1.Specialty = "Ortodonti & Çene Cerrahisi"
			DB.Save(&doc1)
		}

		// Doktor 2: Dr. Zeynep Kaya
		var u2 models.User
		if err := DB.Where("email = ?", "zeynep@dentvisor.com").First(&u2).Error; err != nil {
			u2 = models.User{
				ClinicID:     &clinicID,
				Email:        "zeynep@dentvisor.com",
				PasswordHash: string(hashedPassword),
				Role:         "doctor",
				FirstName:    "Zeynep",
				LastName:     "Kaya",
				Phone:        "0533 111 2233",
			}
			DB.Create(&u2)

			d2 := models.Doctor{
				UserID:         u2.ID,
				ClinicID:       clinicID,
				Title:          "Dt.",
				Specialty:      "Estetik Diş Hekimliği & İmplantoloji",
				CommissionRate: 40.0,
				BaseSalary:     35000.0,
			}
			DB.Create(&d2)
		}

		// Doktor 3: Dr. Mehmet Demir
		var u3 models.User
		if err := DB.Where("email = ?", "mehmet@dentvisor.com").First(&u3).Error; err != nil {
			u3 = models.User{
				ClinicID:     &clinicID,
				Email:        "mehmet@dentvisor.com",
				PasswordHash: string(hashedPassword),
				Role:         "doctor",
				FirstName:    "Mehmet",
				LastName:     "Demir",
				Phone:        "0535 444 5566",
			}
			DB.Create(&u3)

			d3 := models.Doctor{
				UserID:         u3.ID,
				ClinicID:       clinicID,
				Title:          "Dt.",
				Specialty:      "Pedodonti (Çocuk Diş) & Endodonti",
				CommissionRate: 30.0,
				BaseSalary:     30000.0,
			}
			DB.Create(&d3)
		}
	}

	// Doktor ID'lerini alalım
	var doctors []models.Doctor
	DB.Preload("User").Where("clinic_id = ?", clinicID).Find(&doctors)

	var doc1ID, doc2ID, doc3ID string
	if len(doctors) > 0 {
		doc1ID = doctors[0].ID
	}
	if len(doctors) > 1 {
		doc2ID = doctors[1].ID
	} else {
		doc2ID = doc1ID
	}
	if len(doctors) > 2 {
		doc3ID = doctors[2].ID
	} else {
		doc3ID = doc1ID
	}

	// 3. Ünitler (Diş Koltukları)
	var chairCount int64
	DB.Model(&models.Chair{}).Where("clinic_id = ?", clinicID).Count(&chairCount)
	var chairs []models.Chair
	if chairCount == 0 {
		chairs = []models.Chair{
			{ClinicID: clinicID, Name: "Ünit 1 - VIP Cerrahi & İmplant Odası"},
			{ClinicID: clinicID, Name: "Ünit 2 - Estetik & Gülüş Tasarımı"},
			{ClinicID: clinicID, Name: "Ünit 3 - Muayene & Çocuk Odası"},
		}
		DB.Create(&chairs)
	} else {
		DB.Where("clinic_id = ?", clinicID).Find(&chairs)
	}

	var chair1ID, chair2ID, chair3ID string
	if len(chairs) > 0 {
		chair1ID = chairs[0].ID
	}
	if len(chairs) > 1 {
		chair2ID = chairs[1].ID
	} else {
		chair2ID = chair1ID
	}
	if len(chairs) > 2 {
		chair3ID = chairs[2].ID
	} else {
		chair3ID = chair1ID
	}

	// 4. Tedavi Hizmetleri & Fiyat Listesi
	var treatCount int64
	DB.Model(&models.Treatment{}).Where("clinic_id = ?", clinicID).Count(&treatCount)
	var treatments []models.Treatment
	if treatCount < 5 {
		treatments = []models.Treatment{
			{ClinicID: clinicID, Name: "İmplant Tedavisi (Straumann / Osstem)", DefaultPrice: 18000.0, DefaultDuration: 45},
			{ClinicID: clinicID, Name: "Zirkonyum Kaplama / Kron", DefaultPrice: 5500.0, DefaultDuration: 30},
			{ClinicID: clinicID, Name: "E-Max Lamina Porselen", DefaultPrice: 8000.0, DefaultDuration: 40},
			{ClinicID: clinicID, Name: "Kanal Tedavisi (Tek / Çok Kök)", DefaultPrice: 3200.0, DefaultDuration: 60},
			{ClinicID: clinicID, Name: "Estetik Kompozit Dolgu", DefaultPrice: 1800.0, DefaultDuration: 30},
			{ClinicID: clinicID, Name: "Ofis Tipi Lazerle Diş Beyazlatma", DefaultPrice: 6500.0, DefaultDuration: 45},
			{ClinicID: clinicID, Name: "Şeffaf Plak ile Ortodonti (Invisalign)", DefaultPrice: 45000.0, DefaultDuration: 30},
			{ClinicID: clinicID, Name: "20'lik Gömülü Diş Çekimi", DefaultPrice: 4500.0, DefaultDuration: 45},
		}
		DB.Create(&treatments)
	} else {
		DB.Where("clinic_id = ?", clinicID).Find(&treatments)
	}

	var trtImplant, trtZirkon, trtKanal, trtDolgu models.Treatment
	for _, t := range treatments {
		if t.Name == "İmplant Tedavisi (Straumann / Osstem)" {
			trtImplant = t
		}
		if t.Name == "Zirkonyum Kaplama / Kron" {
			trtZirkon = t
		}
		if t.Name == "Kanal Tedavisi (Tek / Çok Kök)" {
			trtKanal = t
		}
		if t.Name == "Estetik Kompozit Dolgu" {
			trtDolgu = t
		}
	}

	// 5. Hastalar (6 Gerçekçi Hasta Profili)
	// Hasta kullanıcı hesapları (Role: patient)
	var patientUser1 models.User
	if err := DB.Where("email = ?", "mert.aksoy@gmail.com").First(&patientUser1).Error; err != nil {
		patientUser1 = models.User{
			ClinicID:     &clinicID,
			Email:        "mert.aksoy@gmail.com",
			PasswordHash: string(hashedPassword),
			Role:         "patient",
			FirstName:    "Mert",
			LastName:     "Aksoy",
			Phone:        "0532 999 1122",
		}
		DB.Create(&patientUser1)
	} else {
		patientUser1.PasswordHash = string(hashedPassword)
		patientUser1.Phone = "0532 999 1122"
		DB.Save(&patientUser1)
	}

	var patientUser2 models.User
	if err := DB.Where("email = ?", "selin.yildiz@hotmail.com").First(&patientUser2).Error; err != nil {
		patientUser2 = models.User{
			ClinicID:     &clinicID,
			Email:        "selin.yildiz@hotmail.com",
			PasswordHash: string(hashedPassword),
			Role:         "patient",
			FirstName:    "Selin",
			LastName:     "Yıldız",
			Phone:        "0535 777 3344",
		}
		DB.Create(&patientUser2)
	} else {
		patientUser2.PasswordHash = string(hashedPassword)
		patientUser2.Phone = "0535 777 3344"
		DB.Save(&patientUser2)
	}

	var patientCount int64
	DB.Model(&models.Patient{}).Where("clinic_id = ?", clinicID).Count(&patientCount)
	var patients []models.Patient
	if patientCount < 5 {
		patients = []models.Patient{
			{
				ClinicID:       clinicID,
				UserID:         &patientUser1.ID,
				FirstName:      "Mert",
				LastName:       "Aksoy",
				Phone:          "0532 999 1122",
				Email:          "mert.aksoy@gmail.com",
				MedicalHistory: "Mide hassasiyeti var, aspirin kullanımı önerilmiyor.",
				Allergies:      "Penisilin alerjisi",
				Balance:        0.0,
			},
			{
				ClinicID:       clinicID,
				UserID:         &patientUser2.ID,
				FirstName:      "Selin",
				LastName:       "Yıldız",
				Phone:          "0535 777 3344",
				Email:          "selin.yildiz@hotmail.com",
				MedicalHistory: "Düzenli tiroid ilacı kullanıyor.",
				Allergies:      "Bilinen alerjisi yok",
				Balance:        12000.0,
			},
			{
				ClinicID:       clinicID,
				FirstName:      "Burak",
				LastName:       "Yılmaz",
				Phone:          "0542 666 5588",
				Email:          "burak.yilmaz@yahoo.com",
				MedicalHistory: "Gece diş sıkma (Bruksizm) şikayeti mevcut.",
				Allergies:      "Lokal anesteziye hafif hassasiyet",
				Balance:        4500.0,
			},
			{
				ClinicID:       clinicID,
				FirstName:      "Canan",
				LastName:       "Demir",
				Phone:          "0533 222 8899",
				Email:          "canan.demir@gmail.com",
				MedicalHistory: "Hipertansiyon kontrol altında.",
				Allergies:      "Bilinen alerjisi yok",
				Balance:        0.0,
			},
			{
				ClinicID:       clinicID,
				FirstName:      "Elif",
				LastName:       "Çelik",
				Phone:          "0530 444 9911",
				Email:          "veli.celik@gmail.com",
				MedicalHistory: "İlk diş muayenesi, çürük yatkınlığı var.",
				Allergies:      "Bilinen alerjisi yok",
				Balance:        0.0,
			},
			{
				ClinicID:       clinicID,
				FirstName:      "Emre",
				LastName:       "Kaya",
				Phone:          "0544 555 7700",
				Email:          "emre.kaya@outlook.com",
				MedicalHistory: "Ön dişlerde travmaya bağlı kırık.",
				Allergies:      "Yok",
				Balance:        2500.0,
			},
		}
		DB.Create(&patients)
	} else {
		DB.Where("clinic_id = ?", clinicID).Find(&patients)
	}

	// 6. Odontogram Kayıtları (Diş Şeması)
	if len(patients) > 0 {
		var recCount int64
		DB.Model(&models.DentalRecord{}).Where("patient_id = ?", patients[0].ID).Count(&recCount)
		if recCount == 0 {
			now := time.Now()
			records := []models.DentalRecord{
				// Mert Aksoy
				{
					PatientID:   patients[0].ID,
					DoctorID:    doc1ID,
					TreatmentID: trtZirkon.ID,
					ToothNumber: "11",
					Status:      "completed",
					Notes:       "Zirkonyum kaplama simante edildi, oklüzyon kontrolü tamam.",
					Date:        now.AddDate(0, 0, -10),
				},
				{
					PatientID:   patients[0].ID,
					DoctorID:    doc1ID,
					TreatmentID: trtDolgu.ID,
					ToothNumber: "21",
					Status:      "completed",
					Notes:       "Mezial estetik kompozit dolgu.",
					Date:        now.AddDate(0, 0, -10),
				},
				{
					PatientID:   patients[0].ID,
					DoctorID:    doc1ID,
					TreatmentID: trtKanal.ID,
					ToothNumber: "46",
					Status:      "planned",
					Notes:       "Derin çürük, kanal tedavisi planlandı.",
					Date:        now,
				},
			}
			if len(patients) > 1 {
				// Selin Yıldız
				records = append(records, []models.DentalRecord{
					{
						PatientID:   patients[1].ID,
						DoctorID:    doc2ID,
						TreatmentID: trtImplant.ID,
						ToothNumber: "16",
						Status:      "completed",
						Notes:       "Osstem 4.0x10mm implant yerleştirildi.",
						Date:        now.AddDate(0, 0, -20),
					},
					{
						PatientID:   patients[1].ID,
						DoctorID:    doc2ID,
						TreatmentID: trtImplant.ID,
						ToothNumber: "26",
						Status:      "planned",
						Notes:       "Sinüs lifting sonrası implant planı.",
						Date:        now,
					},
				}...)
			}
			DB.Create(&records)
		}
	}

	// 7. Tedavi Planları & Teklifler (Faz 2)
	var planCount int64
	DB.Model(&models.TreatmentPlan{}).Where("clinic_id = ?", clinicID).Count(&planCount)
	if planCount == 0 && len(patients) >= 4 {
		tooth11 := 11
		tooth21 := 21
		tooth12 := 12
		tooth22 := 22
		tooth46 := 46
		tooth36 := 36

		// Plan 1: Canan Demir (Estetik Gülüş Tasarımı - Kabul Edildi)
		p1 := models.TreatmentPlan{
			ClinicID:       clinicID,
			PatientID:      patients[3].ID,
			DoctorID:       &doc2ID,
			Title:          "Full Anterior E-Max Gülüş Tasarımı",
			Status:         "accepted",
			Subtotal:       54500.0,
			DiscountRate:   10.0,
			DiscountAmount: 5450.0,
			TotalAmount:    49050.0,
			Notes:          "Fiyatlarımıza KDV ve laboratuvar masrafları dahildir. 2 taksitte ödenecektir.",
			Items: []models.TreatmentPlanItem{
				{ToothNumber: &tooth11, TreatmentName: "E-Max Lamina Porselen", UnitPrice: 8000, Quantity: 1, TotalPrice: 8000, Status: "in_progress"},
				{ToothNumber: &tooth21, TreatmentName: "E-Max Lamina Porselen", UnitPrice: 8000, Quantity: 1, TotalPrice: 8000, Status: "in_progress"},
				{ToothNumber: &tooth12, TreatmentName: "E-Max Lamina Porselen", UnitPrice: 8000, Quantity: 1, TotalPrice: 8000, Status: "in_progress"},
				{ToothNumber: &tooth22, TreatmentName: "E-Max Lamina Porselen", UnitPrice: 8000, Quantity: 1, TotalPrice: 8000, Status: "in_progress"},
				{TreatmentName: "Ofis Tipi Lazerle Diş Beyazlatma", UnitPrice: 6500, Quantity: 1, TotalPrice: 6500, Status: "planned"},
			},
		}
		DB.Create(&p1)

		// Plan 2: Mert Aksoy (İmplant & Kanal - Sunuldu)
		p2 := models.TreatmentPlan{
			ClinicID:       clinicID,
			PatientID:      patients[0].ID,
			DoctorID:       &doc1ID,
			Title:          "Sol ve Sağ Posterior İmplant & Kanal Tedavisi",
			Status:         "presented",
			Subtotal:       39200.0,
			DiscountRate:   0,
			DiscountAmount: 0,
			TotalAmount:    39200.0,
			Notes:          "Teklif 30 gün süreyle geçerlidir.",
			Items: []models.TreatmentPlanItem{
				{ToothNumber: &tooth46, TreatmentName: "Kanal Tedavisi (Çok Kök)", UnitPrice: 3200, Quantity: 1, TotalPrice: 3200, Status: "planned"},
				{ToothNumber: &tooth36, TreatmentName: "İmplant Tedavisi (Straumann)", UnitPrice: 18000, Quantity: 2, TotalPrice: 36000, Status: "planned"},
			},
		}
		DB.Create(&p2)
	}

	// 8. Diş Laboratuvarları & Protez Siparişleri (Faz 3)
	var labCount int64
	DB.Model(&models.Laboratory{}).Where("clinic_id = ?", clinicID).Count(&labCount)
	var labs []models.Laboratory
	if labCount == 0 {
		labs = []models.Laboratory{
			{
				ClinicID:      clinicID,
				Name:          "Estetik Dent Diş Protez Laboratuvarı",
				Phone:         "0212 444 10 20",
				ContactPerson: "Murat Usta (Baş Teknisyen)",
				Address:       "Perpa Ticaret Merkezi B Blok No:142 Şişli / İstanbul",
				Notes:         "E-max ve zirkonyum kaplamalarda uzman.",
			},
			{
				ClinicID:      clinicID,
				Name:          "Tekno Zirkon Cad-Cam Merkezi",
				Phone:         "0216 333 20 30",
				ContactPerson: "Erhan Bey",
				Address:       "Kadıköy Rıhtım Cad. No:18 İstanbul",
				Notes:         "Dijital tarama ve hızlı teslimat.",
			},
		}
		DB.Create(&labs)
	} else {
		DB.Where("clinic_id = ?", clinicID).Find(&labs)
	}

	// Laboratuvar Siparişleri
	var orderCount int64
	DB.Model(&models.LabOrder{}).Where("clinic_id = ?", clinicID).Count(&orderCount)
	if orderCount == 0 && len(labs) >= 2 && len(patients) >= 4 {
		dueDate1 := time.Now().AddDate(0, 0, 3)
		dueDate2 := time.Now().AddDate(0, 0, -2) // Gecikmiş prova örneği
		dueDate3 := time.Now().AddDate(0, 0, 7)

		orders := []models.LabOrder{
			{
				ClinicID:     clinicID,
				PatientID:    patients[3].ID, // Canan Demir
				DoctorID:     &doc2ID,
				LaboratoryID: labs[0].ID,
				WorkType:     "E-Max Lamina Porselen",
				ToothNumbers: "11, 12, 21, 22",
				ShadeColor:   "BL2 (Bleach White)",
				Status:       "try_in", // Provada
				SentDate:     time.Now().AddDate(0, 0, -4),
				DueDate:      &dueDate1,
				Cost:         7200.0,
				IsPaid:       false,
				Notes:        "Gülüş hattı yüksek, insizal şeffaflık belirgin olsun.",
			},
			{
				ClinicID:     clinicID,
				PatientID:    patients[0].ID, // Mert Aksoy
				DoctorID:     &doc1ID,
				LaboratoryID: labs[1].ID,
				WorkType:     "Zirkonyum Kaplama",
				ToothNumbers: "11",
				ShadeColor:   "A1",
				Status:       "completed", // Teslim Alındı
				SentDate:     time.Now().AddDate(0, 0, -12),
				DueDate:      &dueDate2,
				Cost:         1800.0,
				IsPaid:       true,
				Notes:        "Anterior oklüzyon kontrol edildi.",
			},
			{
				ClinicID:     clinicID,
				PatientID:    patients[1].ID, // Selin Yıldız
				DoctorID:     &doc2ID,
				LaboratoryID: labs[0].ID,
				WorkType:     "İmplant Üstü Zirkonyum Abutment & Kron",
				ToothNumbers: "16",
				ShadeColor:   "A2",
				Status:       "in_lab", // Laboratuvarda
				SentDate:     time.Now().AddDate(0, 0, -2),
				DueDate:      &dueDate3,
				Cost:         2500.0,
				IsPaid:       false,
				Notes:        "Açılı abutment kullanılacak.",
			},
			{
				ClinicID:     clinicID,
				PatientID:    patients[2].ID, // Burak Yılmaz
				DoctorID:     &doc1ID,
				LaboratoryID: labs[1].ID,
				WorkType:     "Gece Plağı (Bruksizm)",
				ToothNumbers: "Genel Üst Çene",
				ShadeColor:   "Şeffaf Sert/Yumuşak",
				Status:       "sent", // Ölçü Gönderildi
				SentDate:     time.Now(),
				DueDate:      &dueDate1,
				Cost:         850.0,
				IsPaid:       false,
				Notes:        "3mm çift katmanlı sert-yumuşak plak.",
			},
		}
		DB.Create(&orders)
	}

	// 9. WhatsApp Business & AI Chatbot Canlı Verileri (Faz 4)
	var waSetting models.WhatsappSetting
	if err := DB.Where("clinic_id = ?", clinicID).First(&waSetting).Error; err != nil {
		waSetting = models.WhatsappSetting{
			ClinicID:            clinicID,
			PhoneNumber:         "+90 532 555 4040",
			Status:              "connected",
			IsAiEnabled:         true,
			AiBotName:           "Dentvisör Akıllı Klinik Asistanı",
			AiInstructions:      "Kliniğimiz Nişantaşı Valikonağı Caddesi'ndedir. Hafta içi ve cumartesi 09:00 - 20:00 saatleri arasında açığız. Estetik diş hekimliği, implant ve ortodontide uzmanız. Randevu taleplerini nazikçe al ve hastaya yardımcı ol.",
			AutoReminderEnabled: true,
			ReminderHoursBefore: 24,
			AiReminderTone:      "warm_professional",
			PostCareEnabled:     true,
			GoogleReviewLink:    "https://g.page/r/dentvisor-nisantasi/review",
		}
		DB.Create(&waSetting)
	} else {
		waSetting.Status = "connected"
		waSetting.PhoneNumber = "+90 532 555 4040"
		DB.Save(&waSetting)
	}

	// WhatsApp Mesajlaşma Günlüğü
	var msgCount int64
	DB.Model(&models.WhatsappMessageLog{}).Where("clinic_id = ?", clinicID).Count(&msgCount)
	if msgCount == 0 {
		now := time.Now()
		logs := []models.WhatsappMessageLog{
			{
				ClinicID:      clinicID,
				Direction:     "incoming",
				Phone:         "+90 532 999 1122",
				SenderName:    "Mert Aksoy",
				Message:       "İyi günler, Zirkonyum kaplama fiyatlarınız hakkında bilgi alabilir miyim?",
				IsAiGenerated: false,
				Status:        "received",
				CreatedAt:     now.Add(-time.Hour * 3),
			},
			{
				ClinicID:      clinicID,
				Direction:     "outgoing",
				Phone:         "+90 532 999 1122",
				SenderName:    "Dentvisör Akıllı Klinik Asistanı",
				Message:       "Merhaba Mert Bey! Özel Dentvisör Polikliniği'ne hoş geldiniz. ✨ Kliniğimizde zirkonyum kaplama tedavilerimiz tek üye ₺5.500 olarak uygulanmaktadır. Kesin tedavi planı için hekimimizin ücretsiz ön muayenesine davet etmek isteriz. Randevu planlayalım mı?",
				IsAiGenerated: true,
				Status:        "delivered",
				CreatedAt:     now.Add(-time.Hour*3 + time.Second*5),
			},
			{
				ClinicID:      clinicID,
				Direction:     "incoming",
				Phone:         "+90 535 777 3344",
				SenderName:    "Selin Yıldız",
				Message:       "Yarın saat 14:00 için Dr. Zeynep Hanım'a kontrol randevum vardı, teyit etmek istedim.",
				IsAiGenerated: false,
				Status:        "received",
				CreatedAt:     now.Add(-time.Hour * 1),
			},
			{
				ClinicID:      clinicID,
				Direction:     "outgoing",
				Phone:         "+90 535 777 3344",
				SenderName:    "Dentvisör Akıllı Klinik Asistanı",
				Message:       "Merhaba Selin Hanım! 🦷 Randevunuzu kontrol ettim, yarın saat 14:00'te Dr. Zeynep Kaya ile implant kontrol randevunuz sistemimizde onaylıdır. Randevudan 10 dakika önce kliniğimizde olmanızı rica ederiz. Sağlıklı günler!",
				IsAiGenerated: true,
				Status:        "read",
				CreatedAt:     now.Add(-time.Hour*1 + time.Second*8),
			},
		}
		DB.Create(&logs)
	}

	// 10. Randevular
	var apptCount int64
	DB.Model(&models.Appointment{}).Where("clinic_id = ?", clinicID).Count(&apptCount)
	if apptCount < 4 && len(patients) >= 4 {
		now := time.Now()
		appts := []models.Appointment{
			{
				ClinicID:    clinicID,
				PatientID:   patients[0].ID,
				DoctorID:    doc1ID,
				TreatmentID: trtZirkon.ID,
				ChairID:     chair1ID,
				StartTime:   now.AddDate(0, 0, -1).Truncate(time.Hour).Add(10 * time.Hour),
				EndTime:     now.AddDate(0, 0, -1).Truncate(time.Hour).Add(11 * time.Hour),
				Status:      "completed",
				Notes:       "Zirkonyum prova yapıldı.",
			},
			{
				ClinicID:    clinicID,
				PatientID:   patients[1].ID,
				DoctorID:    doc2ID,
				TreatmentID: trtImplant.ID,
				ChairID:     chair1ID,
				StartTime:   now.Truncate(time.Hour).Add(14 * time.Hour),
				EndTime:     now.Truncate(time.Hour).Add(15 * time.Hour),
				Status:      "confirmed",
				Notes:       "İmplant operasyonu.",
			},
			{
				ClinicID:    clinicID,
				PatientID:   patients[3].ID,
				DoctorID:    doc2ID,
				TreatmentID: trtDolgu.ID,
				ChairID:     chair2ID,
				StartTime:   now.AddDate(0, 0, 1).Truncate(time.Hour).Add(16 * time.Hour),
				EndTime:     now.AddDate(0, 0, 1).Truncate(time.Hour).Add(17 * time.Hour),
				Status:      "scheduled",
				Notes:       "E-Max estetik prova.",
			},
			{
				ClinicID:    clinicID,
				PatientID:   patients[4].ID, // Elif Çelik (Pedodonti)
				DoctorID:    doc3ID,
				TreatmentID: trtDolgu.ID,
				ChairID:     chair3ID,
				StartTime:   now.AddDate(0, 0, 2).Truncate(time.Hour).Add(11 * time.Hour),
				EndTime:     now.AddDate(0, 0, 2).Truncate(time.Hour).Add(12 * time.Hour),
				Status:      "scheduled",
				Notes:       "İlk muayene ve fissür örtücü.",
			},
		}
		DB.Create(&appts)
	}

	// 11. Hekim Hakediş & Bordro Ödemeleri (Faz 5)
	var payoutCount int64
	DB.Model(&models.DoctorPayout{}).Where("clinic_id = ?", clinicID).Count(&payoutCount)
	if payoutCount == 0 {
		paidDate := time.Now().AddDate(0, -1, 0)
		payouts := []models.DoctorPayout{
			{
				ClinicID:       clinicID,
				DoctorID:       doc1ID,
				PeriodMonth:    int(time.Now().Month()) - 1,
				PeriodYear:     time.Now().Year(),
				TotalRevenue:   120000.0,
				CommissionRate: 35.0,
				CommissionEarn: 42000.0,
				BaseSalary:     40000.0,
				BonusAmount:    5000.0,
				Deductions:     0.0,
				NetPayout:      87000.0,
				Status:         "paid",
				PaidAt:         &paidDate,
				PaymentMethod:  "bank_transfer",
				Notes:          "Geçen ay hakediş ve başarı primi ödendi.",
			},
			{
				ClinicID:       clinicID,
				DoctorID:       doc2ID,
				PeriodMonth:    int(time.Now().Month()) - 1,
				PeriodYear:     time.Now().Year(),
				TotalRevenue:   165000.0,
				CommissionRate: 40.0,
				CommissionEarn: 66000.0,
				BaseSalary:     35000.0,
				BonusAmount:    0.0,
				Deductions:     1500.0,
				NetPayout:      99500.0,
				Status:         "paid",
				PaidAt:         &paidDate,
				PaymentMethod:  "bank_transfer",
				Notes:          "İmplant ve E-max vakaları tamamlandı.",
			},
		}
		DB.Create(&payouts)
	}

	// 12. Çoklu Klinik Örneği (2. Klinik: Özel Dentart Kadıköy) & Mert Aksoy Birleşik Dosyası
	var clinic2 models.Clinic
	cityID2 := uint(34)     // İstanbul
	districtID2 := uint(1651) // Kadıköy
	err = DB.Where("name = ?", "Özel Dentart Ağız ve Diş Sağlığı Polikliniği").First(&clinic2).Error
	if err != nil {
		clinic2 = models.Clinic{
			Name:         "Özel Dentart Ağız ve Diş Sağlığı Polikliniği",
			Phone:        "0216 444 34 34",
			Address:      "Bağdat Cad. No:128/A, Kadıköy / İstanbul",
			AboutText:    "Dentart Polikliniği, çene cerrahisi ve implantolojide 20 yıllık tecrübeyle Kadıköy Bağdat Caddesi'nde hizmet vermektedir.",
			WorkingHours: "Pzt - Cmt: 09:00 - 19:30",
			IsActive:     true,
			CityID:       &cityID2,
			DistrictID:   &districtID2,
		}
		DB.Create(&clinic2)

		// 2. Klinik Doktoru
		uDoc2 := models.User{
			ClinicID:     &clinic2.ID,
			Email:        "ali.veli@dentart.com",
			PasswordHash: string(hashedPassword),
			Role:         "doctor",
			FirstName:    "Ali",
			LastName:     "Veli",
			Phone:        "0532 888 7766",
		}
		DB.Create(&uDoc2)

		docKadıkoy := models.Doctor{
			UserID:         uDoc2.ID,
			ClinicID:       clinic2.ID,
			Title:          "Prof. Dt.",
			Specialty:      "Ağız, Diş ve Çene Cerrahisi",
			CommissionRate: 40.0,
			BaseSalary:     50000.0,
		}
		DB.Create(&docKadıkoy)

		// 2. Klinik Tedavi Hizmeti
		trtKadıkoy := models.Treatment{
			ClinicID:        clinic2.ID,
			Name:            "İmplant Tedavisi (Nobel Biocare)",
			DefaultPrice:    18000.0,
			DefaultDuration: 45,
		}
		DB.Create(&trtKadıkoy)

		// Mert Aksoy'un 2. Klinikteki Hasta Dosyası
		p2Kadıkoy := models.Patient{
			ClinicID:       clinic2.ID,
			UserID:         &patientUser1.ID,
			FirstName:      "Mert",
			LastName:       "Aksoy",
			Phone:          "0532 999 1122",
			Email:          "mert.aksoy@gmail.com",
			MedicalHistory: "2025 yılında sol alt çene #36 nolu dişe implant uygulandı.",
			Allergies:      "Penisilin alerjisi",
			Balance:        0.0,
		}
		DB.Create(&p2Kadıkoy)

		// 2. Klinikteki Diş Kaydı (#36 İmplant)
		recKadıkoy := models.DentalRecord{
			PatientID:   p2Kadıkoy.ID,
			DoctorID:    docKadıkoy.ID,
			TreatmentID: trtKadıkoy.ID,
			ToothNumber: "36",
			Status:      "completed",
			Notes:       "Nobel Biocare 4.3x11.5mm implant yerleştirildi ve zirkonyum kron takıldı.",
			Date:        time.Now().AddDate(-1, -2, 0), // 1 yıl 2 ay önce
		}
		DB.Create(&recKadıkoy)

		// 2. Klinikteki Geçmiş Randevu
		apptKadıkoy := models.Appointment{
			ClinicID:    clinic2.ID,
			PatientID:   p2Kadıkoy.ID,
			DoctorID:    docKadıkoy.ID,
			TreatmentID: trtKadıkoy.ID,
			StartTime:   time.Now().AddDate(0, 0, 7).Truncate(time.Hour).Add(15 * time.Hour),
			EndTime:     time.Now().AddDate(0, 0, 7).Truncate(time.Hour).Add(16 * time.Hour),
			Status:      "confirmed",
			Notes:       "Yıllık rutin implant çevre doku kontrolü.",
		}
		DB.Create(&apptKadıkoy)

		// 2. Klinikteki Tamamlanmış Tedavi Planı
		tooth36 := 36
		planKadıkoy := models.TreatmentPlan{
			ClinicID:       clinic2.ID,
			PatientID:      p2Kadıkoy.ID,
			DoctorID:       &docKadıkoy.ID,
			Title:          "Sol Alt 1. Molar İmplant & Zirkonyum Kron",
			Status:         "accepted",
			Subtotal:       18000.0,
			DiscountRate:   0,
			DiscountAmount: 0,
			TotalAmount:    18000.0,
			Notes:          "Tedavi tamamlandı ve garanti belgesi teslim edildi.",
			Items: []models.TreatmentPlanItem{
				{ToothNumber: &tooth36, TreatmentName: "İmplant Tedavisi (Nobel Biocare)", UnitPrice: 18000, Quantity: 1, TotalPrice: 18000, Status: "completed"},
			},
		}
		DB.Create(&planKadıkoy)

		// KVKK Konsültasyon İzni (Dentvisör Nişantaşı için onaylanmış izin)
		consent := models.PatientConsent{
			PatientID:      patients[0].ID,
			TargetClinicID: clinicID,
			Status:         "approved",
			ExpiresAt:      time.Now().AddDate(0, 1, 0),
		}
		DB.Create(&consent)
	}

	log.Printf("✅ admin@test.com (%s) ve çoklu klinik demo verileri başarıyla yüklendi!", clinic.Name)
}
