package repositories

import (
	"dentvisor-backend/internal/models"
	"dentvisor-backend/pkg/database"
)

type UserRepository struct{}

func (r *UserRepository) CreateUser(user *models.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmailOrPhone(identifier string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("LOWER(email) = LOWER(?) OR phone = ?", identifier, identifier).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
