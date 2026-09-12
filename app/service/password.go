package service

import "unicode"

// IsStrongPassword memeriksa apakah password memenuhi aturan keamanan.
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}
