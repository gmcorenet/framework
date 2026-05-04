package security

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	BCryptCost = 10
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
	return string(bytes), err
}

func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func CheckPasswordStrength(password string) PasswordStrength {
	score := 0
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if hasUpperCase(password) {
		score++
	}
	if hasLowerCase(password) {
		score++
	}
	if hasDigit(password) {
		score++
	}
	if hasSpecialChar(password) {
		score++
	}
	return PasswordStrength(score)
}

type PasswordStrength int

const (
	PasswordStrengthVeryWeak PasswordStrength = iota
	PasswordStrengthWeak
	PasswordStrengthMedium
	PasswordStrengthStrong
	PasswordStrengthVeryStrong
)

func (s PasswordStrength) Score() int {
	return int(s)
}

func (s PasswordStrength) IsStrong(minStrength int) bool {
	return int(s) >= minStrength
}

func hasUpperCase(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

func hasLowerCase(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

func hasSpecialChar(s string) bool {
	special := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	for _, c := range s {
		for _, sc := range special {
			if c == sc {
				return true
			}
		}
	}
	return false
}
