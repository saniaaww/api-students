package service

import (
	"context"
	"errors"

	"api-students/app/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidLogin = errors.New("username atau password salah")

func LoginUser(
	ctx context.Context,
	repo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	username string,
	password string,
	jwtSecret string,
) (string, string, error) {

	user, err := repo.FindByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	// Username tidak ditemukan
	if user == nil {
		return "", "", ErrInvalidLogin
	}

	// Password tidak cocok
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	); err != nil {
		return "", "", ErrInvalidLogin
	}

	// Membuat access token JWT
	accessToken, err := GenerateAccessToken(user, jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Membuat refresh token
	refreshToken, refreshTokenHash, err := GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Menyimpan hash refresh token ke database
	err = refreshRepo.Create(
		ctx,
		user.ID,
		refreshTokenHash,
		RefreshTokenExpiresAt(),
	)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}