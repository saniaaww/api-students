package service

import (
	"context"
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrUsernameExists = errors.New("username sudah digunakan")
	ErrEmailExists    = errors.New("email sudah digunakan")
)

func RegisterUser(
	ctx context.Context,
	repo repository.UserRepository,
	req model.RegisterRequest,
) (*model.User, error) {

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)

	existingUsername, err := repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if existingUsername != nil {
		return nil, ErrUsernameExists
	}

	existingEmail, err := repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingEmail != nil {
		return nil, ErrEmailExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcryptCost,
	)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: string(passwordHash),
		Role:     "user",
	}

	return repo.Create(ctx, user)
}