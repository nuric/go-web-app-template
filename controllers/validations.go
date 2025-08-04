package controllers

import (
	"errors"
	"regexp"
	"strings"
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("invalid email format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 ||
		!regexp.MustCompile(`[a-z]`).MatchString(password) ||
		!regexp.MustCompile(`[A-Z]`).MatchString(password) ||
		!regexp.MustCompile(`\d`).MatchString(password) ||
		!regexp.MustCompile(`[@$!%*?&=]`).MatchString(password) {
		return errors.New("password must be at least 8 characters long, contain at least one lowercase letter, one uppercase letter, one digit, and one special character")
	}
	return nil
}

func ValidateToken(token string) error {
	if len(token) < 6 {
		return errors.New("token must be at least 6 characters long")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, token); !matched {
		return errors.New("invalid token format")
	}
	return nil
}
