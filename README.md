# Dentvisör Backend

Go (Gin) + PostgreSQL ile yazılmış REST API sunucusu.

---

## 🚀 Hızlı Başlangıç

### Gereksinimler

| Araç      | Versiyon | Kontrol            |
| ---------- | -------- | ------------------ |
| Go         | 1.21+    | `go version`     |
| PostgreSQL | 14+      | `psql --version` |

---

## 1. Veritabanı Kurulumu

PostgreSQL'de veritabanı oluşturun:

```bash
psql -U postgres -c "CREATE DATABASE dentvisor;"
```

> Eğer PostgreSQL kurulu değilse: [https://postgresql.org/download/macosx/](https://postgresql.org/download/macosx/)
> macOS için: `brew install postgresql@16 && brew services start postgresql@16`

---

## 2. Ortam Değişkenleri (`.env`)

`backend/` dizininde `.env` dosyası **zaten mevcut**. İçeriği:

```env
PORT=8888
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dentvisor
DB_PORT=5432
JWT_SECRET=super-secret-dentvisor-key
```

Kendi PostgreSQL şifreniz farklıysa `DB_PASSWORD` satırını güncelleyin.

---

## 3. Backend'i Çalıştırma

```bash
# backend/ dizinine girin
cd backend

# Bağımlılıkları indir (ilk seferinde)
go mod download

# Sunucuyu başlat
go run cmd/api/main.go
```

Başarılı çıktı şöyle görünür:

```
2026/08/17 12:00:00 PostgreSQL veritabanına başarıyla bağlanıldı!
2026/08/17 12:00:00 AutoMigrate başarıyla tamamlandı
2026/08/17 12:00:00 Gerçek İl ve İlçe verileri okunuyor (Yerel dosyalardan)...
2026/08/17 12:00:01 Tüm 81 il ve ilçeleri veritabanına başarıyla eklendi!
[GIN-debug] Listening and serving HTTP on :8888
```

> **Not:** Sunucu başlarken il/ilçe verileri otomatik olarak veritabanına yüklenir.

---

## 4. Çalıştığını Doğrulama

```bash
curl http://localhost:8888/api/health
# → {"status":"ok"}

curl http://localhost:8888/api/locations/cities | head -c 100
# → [{"id":1,"name":"ADANA",...
```

---

## 5. Derleme (Opsiyonel)

Tek bir binary olarak derlemek için:

```bash
cd backend
go build -o api cmd/api/main.go
./api
```

---

## 📁 Proje Yapısı

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # Giriş noktası, router tanımları
├── internal/
│   ├── config/              # Ortam değişkeni okuma
│   ├── handlers/            # HTTP handler'lar (Auth, Patients, Appointments...)
│   ├── middleware/          # JWT auth middleware
│   ├── models/              # GORM modelleri (User, Clinic, Patient...)
│   ├── repositories/        # Veritabanı sorguları
│   └── services/            # İş mantığı katmanı
├── pkg/
│   └── database/
│       ├── db.go            # PostgreSQL bağlantısı
│       └── seed.go          # İl/İlçe veri yükleme
├── db/
│   └── seed/
│       └── data/
│           ├── il.json      # 81 il
│           └── ilce.json    # Tüm ilçeler
├── .env                     # Ortam değişkenleri
└── go.mod
```

---

## 🔌 API Endpoint'leri

### Public (Token gerekmez)

| Method | URL                                     | Açıklama          |
| ------ | --------------------------------------- | ------------------- |
| GET    | `/api/health`                         | Sağlık kontrolü  |
| POST   | `/api/auth/register`                  | Yeni klinik kaydı  |
| POST   | `/api/auth/login`                     | Giriş, JWT döner  |
| GET    | `/api/locations/cities`               | Tüm iller          |
| GET    | `/api/locations/cities/:id/districts` | İle göre ilçeler |

### Protected (Bearer Token gerekir)

| Method | URL                                          | Açıklama                   |
| ------ | -------------------------------------------- | ---------------------------- |
| GET    | `/api/protected/settings/clinic`           | Klinik bilgileri             |
| PUT    | `/api/protected/settings/clinic`           | Klinik bilgilerini güncelle |
| GET    | `/api/protected/settings/doctors`          | Doktorlar                    |
| GET    | `/api/protected/settings/treatments`       | Tedaviler                    |
| GET    | `/api/protected/settings/chairs`           | Koltuklar                    |
| GET    | `/api/protected/patients`                  | Hasta listesi                |
| POST   | `/api/protected/patients`                  | Yeni hasta ekle              |
| GET    | `/api/protected/patients/:id`              | Hasta detayı                |
| GET    | `/api/protected/appointments`              | Randevular                   |
| POST   | `/api/protected/appointments`              | Yeni randevu                 |
| PUT    | `/api/protected/appointments/:id/status`   | Randevu durumu güncelle     |
| GET    | `/api/protected/admin/reports/performance` | Hekim performans raporu      |

---

## 🔐 Kayıt & Giriş Örneği

```bash
# Kayıt ol
curl -X POST http://localhost:8888/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Ahmet",
    "last_name": "Yılmaz",
    "clinic_name": "Gülüş Diş Kliniği",
    "email": "ahmet@klinik.com",
    "password": "sifre123"
  }'

# Giriş yap (token al)
curl -X POST http://localhost:8888/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "ahmet@klinik.com", "password": "sifre123"}'

# Token ile protected endpoint'e istek at
curl http://localhost:8888/api/protected/patients \
  -H "Authorization: Bearer <TOKEN>"
```

---

## ⚠️ Sık Karşılaşılan Sorunlar

| Hata                                    | Çözüm                                                        |
| --------------------------------------- | --------------------------------------------------------------- |
| `connection refused` (DB)             | PostgreSQL çalışmıyor.`brew services start postgresql@16` |
| `database "dentvisor" does not exist` | `psql -U postgres -c "CREATE DATABASE dentvisor;"`            |
| `password authentication failed`      | `.env` dosyasında `DB_PASSWORD` değerini güncelle        |
| `port 8888 already in use`            | `lsof -ti:8888 \| xargs kill`                                  |
