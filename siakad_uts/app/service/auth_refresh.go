package service

import (
	"context"
	"errors"

	"api-students/app/repository"
)

var ErrInvalidRefreshToken = errors.New("refresh token tidak valid")

func RefreshAccessToken(
	ctx context.Context,
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	refreshToken string,
	jwtSecret string,
) (string, string, error) {

	// Hash refresh token yang diberikan client.
	hash := HashRefreshToken(refreshToken)

	// Cari refresh token yang masih aktif di database.
	userID, err := refreshRepo.FindActive(ctx, hash)
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}

	// Ambil data user berdasarkan ID.
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return "", "", ErrInvalidRefreshToken
	}

	// Revoke refresh token lama.
	if err := refreshRepo.Revoke(ctx, hash); err != nil {
		return "", "", err
	}

	// Buat access token baru.
	accessToken, err := GenerateAccessToken(user, jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Buat refresh token baru.
	newRefreshToken, newRefreshTokenHash, err := GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Simpan refresh token baru.
	if err := refreshRepo.Create(
		ctx,
		user.ID,
		newRefreshTokenHash,
		RefreshTokenExpiresAt(),
	); err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
