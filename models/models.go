package models

import "time"

type User struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	UpstreamID string     `gorm:"uniqueIndex" json:"upstream_id"`
	CreatedAt  time.Time  `json:"created_at"`
	AuthCodes  []AuthCode `gorm:"foreignKey:UserID" json:"-"`
	Caches     []Cache    `gorm:"foreignKey:UserID" json:"-"`
}

type AuthCode struct {
	Code      string    `gorm:"primaryKey" json:"code"`
	UserID    int64     `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Cache struct {
	PhpHash   string `gorm:"primaryKey" json:"php_hash"`
	UserID    int64  `json:"user_id"`
	User      User   `gorm:"foreignKey:UserID" json:"-"`
	ExpiresAt int64  `json:"expires_at"`
}
