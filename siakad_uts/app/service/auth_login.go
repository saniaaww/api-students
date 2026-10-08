package service

import (
	"context"
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidLogin = errors.New("email atau password salah")

// Access token berlaku selama 15 menit.
// 15 menit = 900 detik.
const accessTokenExpiresIn int64 = 900

func LoginUser(
	ctx context.Context,
	repo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	email string,
	password string,
	jwtSecret string,
) (string, int64, *model.User, error) {

	// Bersihkan email
	email = strings.TrimSpace(
		strings.ToLower(email),
	)

	// ==========================================
	// 1. CARI USER BERDASARKAN EMAIL
	// ==========================================

	user, err := repo.FindByEmail(
		ctx,
		email,
	)

	if err != nil {
		return "", 0, nil, err
	}

	// Email tidak ditemukan
	if user == nil {
		return "", 0, nil, ErrInvalidLogin
	}

	// ==========================================
	// 2. CEK PASSWORD
	// ==========================================

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return "", 0, nil, ErrInvalidLogin
	}

	// ==========================================
	// 3. BUAT ACCESS TOKEN JWT
	// ==========================================

	accessToken, err := GenerateAccessToken(
		user,
		jwtSecret,
	)

	if err != nil {
		return "", 0, nil, err
	}

	// ==========================================
	// 4. RETURN
	// ==========================================

	return accessToken, accessTokenExpiresIn, user, nil
}