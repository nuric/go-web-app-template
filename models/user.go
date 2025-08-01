package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email         string `gorm:"uniqueIndex;not null"`
	Password      string `gorm:"not null"`
	Role          string `gorm:"default:'basic'"`
	EmailVerified bool   `gorm:"default:false"`
}

type Token struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	Purpose   string    `gorm:"not null"` // e.g., "password_reset", "email_verification"
	ExpiresAt time.Time `gorm:"not null"`
}
