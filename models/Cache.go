package models

import "time"

type AuthCode struct {
	Code      string    `gorm:"primaryKey" json:"code"`
	UserID    int64     `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
