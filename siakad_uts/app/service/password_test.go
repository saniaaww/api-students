package service

import "testing"

func TestIsStrongPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "password kuat",
			password: "Password1!",
			expected: true,
		},
		{
			name:     "tidak ada huruf besar",
			password: "password1!",
			expected: false,
		},
		{
			name:     "tidak ada angka",
			password: "Password!",
			expected: false,
		},
		{
			name:     "kurang dari 8 karakter",
			password: "Pass1!",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStrongPassword(tt.password)

			if result != tt.expected {
				t.Errorf(
					"IsStrongPassword(%q) = %v, expected %v",
					tt.password,
					result,
					tt.expected,
				)
			}
		})
	}
}
