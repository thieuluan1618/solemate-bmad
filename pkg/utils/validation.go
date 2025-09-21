package utils

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidPassword(password string) bool {
	return len(password) >= 8
}

func IsValidPhoneNumber(phone string) bool {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	if strings.HasPrefix(cleaned, "+") {
		cleaned = cleaned[1:]
	}

	phoneRegex := regexp.MustCompile(`^\d{10,15}$`)
	return phoneRegex.MatchString(cleaned)
}

func SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

func IsValidRole(role string) bool {
	validRoles := []string{"customer", "admin", "manager"}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}
