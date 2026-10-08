package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const refreshTokenTTL = 24 * time.Hour

func GenerateRefreshToken() (string, string, error) {
	rawToken := make([]byte, 32)

	if _, err := rand.Read(rawToken); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(rawToken)

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	return token, tokenHash, nil
}

func RefreshTokenExpiresAt() time.Time {
	return time.Now().Add(refreshTokenTTL)
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
