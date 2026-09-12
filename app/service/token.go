package service

import (
	"errors"
	"time"

	"api-students/app/model"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("token tidak valid")

const accessTokenTTL = 15 * time.Minute

func GenerateAccessToken(user *model.User, secret string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"iat":      now.Unix(),
		"exp":      now.Add(accessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
