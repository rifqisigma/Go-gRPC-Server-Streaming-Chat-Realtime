package repository

import (
	"chat_api/entity"
	"context"

	"gorm.io/gorm"
)

type AuthRepository interface {
	Login(ctx context.Context, user *entity.User) (string, uint, error)
	Register(ctx context.Context, user *entity.User) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepo(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) Login(ctx context.Context, user *entity.User) (string, uint, error) {

	var finduser entity.User
	if err := r.db.Model(&entity.User{}).Where("email = ?", user.Email).First(&finduser).Error; err != nil {
		return "", 0, err

	}

	return finduser.Password, finduser.ID, nil
}

func (r *authRepository) Register(ctx context.Context, user *entity.User) error {
	createuser := entity.User{
		Email:    user.Email,
		Password: user.Password,
		Username: user.Username,
	}

	if err := r.db.Create(&createuser).Error; err != nil {
		return err
	}

	return nil
}
