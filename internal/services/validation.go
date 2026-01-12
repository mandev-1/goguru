package services

import (
	"errors"
	"regexp"
)

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func ValidateRegistration(username, email, password, confirm string) error {
	if len(username) < 3 || len(username) > 20 {
		return errors.New("Username must be 3-20 characters")
	}
	for _, c := range username {
		if !(c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return errors.New("Username may contain letters, numbers, underscore")
		}
	}
	if !emailRe.MatchString(email) {
		return errors.New("Invalid email address")
	}
	if len(password) < 8 {
		return errors.New("Password must be at least 8 characters")
	}
	if password != confirm {
		return errors.New("Passwords do not match")
	}
	return nil
}

func SanitizeFilename(s string) string {
	result := ""
	for _, c := range s {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-' {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32) // to lowercase
		} else {
			result += "-"
		}
	}
	// Trim leading/trailing dashes and underscores
	for len(result) > 0 && (result[0] == '-' || result[0] == '_') {
		result = result[1:]
	}
	for len(result) > 0 && (result[len(result)-1] == '-' || result[len(result)-1] == '_') {
		result = result[:len(result)-1]
	}
	if result == "" {
		result = "asset"
	}
	return result
}
