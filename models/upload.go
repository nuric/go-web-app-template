package models

import (
	"gorm.io/gorm"
)

type Upload struct {
	gorm.Model
	GUID     string `gorm:"uniqueIndex;not null"`
	UserID   uint
	FileName string `gorm:"not null"`
	Size     int64  `gorm:"not null"`
	Mime     string `gorm:"not null"`
}
