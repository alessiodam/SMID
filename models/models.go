package models

import (
	"time"
)

type Cache struct {
	PhpHash   string    `gorm:"primaryKey" json:"php_hash"`
	UserID    int64     `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
