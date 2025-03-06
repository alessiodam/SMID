package models

import "time"

type User struct {
	ID         string `gorm:"primaryKey"`
	UpstreamID string `gorm:"uniqueIndex"`
	CreatedAt  time.Time
}

type AuthCode struct {
	Code      string `gorm:"primaryKey"`
	UserID    string
	CreatedAt time.Time
}

type Cache struct {
	PhpHash   string `gorm:"primaryKey"`
	UserID    string
	ExpiresAt int64
}
