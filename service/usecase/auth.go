package usecase

import (
	"chat_api/entity"
	"chat_api/service/repository"
	"chat_api/utils/helper"
	"context"
	"errors"
	"fmt"
)

type AuthUsecase interface {
	Login(ctx context.Context, user *entity.User) (string, error)
	Register(ctx context.Context, user *entity.User) error
}

type authUsecase struct {
	authRepo repository.AuthRepository
}

func NewAuthUsecase(authRepo repository.AuthRepository) AuthUsecase {
	return &authUsecase{authRepo}
}

func (u *authUsecase) Login(ctx context.Context, user *entity.User) (string, error) {
	pw, id, err := u.authRepo.Login(ctx, user)
	if err != nil {
		return "", nil
	}
	fmt.Println(pw)
	valid := helper.ComparePassword(pw, user.Password)
	if !valid {
		return "", errors.New("password dan email tidak cocok")
	}

	token, err := helper.GenerateJWTLogin(id, user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *authUsecase) Register(ctx context.Context, user *entity.User) error {
	pw, err := helper.HashPassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = pw
	return u.authRepo.Register(ctx, user)
}
