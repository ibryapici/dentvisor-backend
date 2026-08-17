package services

import (
	"errors"
	"os"
	"time"

	"dentvisor-backend/internal/models"
	"dentvisor-backend/internal/repositories"
	"dentvisor-backend/pkg/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: &repositories.UserRepository{},
	}
}

func (s *AuthService) RegisterDoctor(req models.RegisterRequest) (*models.User, error) {
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("bu e-posta adresi zaten kullanımda")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var createdUser models.User
	
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Create Clinic
		clinic := models.Clinic{
			Name: req.ClinicName,
		}
		if err := tx.Create(&clinic).Error; err != nil {
			return err
		}

		// 2. Create User (Admin/Doctor)
		user := models.User{
			ClinicID:     &clinic.ID,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			Role:         "doctor", // Kayıt olan ilk kullanıcı hekim/admin rolündedir
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Phone:        req.Phone,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		createdUser = user

		// 3. Create Doctor Profile
		doctor := models.Doctor{
			UserID:   user.ID,
			ClinicID: clinic.ID,
		}
		if err := tx.Create(&doctor).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &createdUser, nil
}

func (s *AuthService) Login(req models.LoginRequest) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", nil, errors.New("geçersiz e-posta veya şifre")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", nil, errors.New("geçersiz e-posta veya şifre")
	}

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 72).Unix(),
	}
	if user.ClinicID != nil {
		claims["clinic_id"] = *user.ClinicID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}
