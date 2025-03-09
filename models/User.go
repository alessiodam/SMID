package models

import (
	"github.com/go-playground/validator/v10"
	"time"
)

type User struct {
	ID         int64  `gorm:"primaryKey" json:"id"`
	UpstreamID string `gorm:"uniqueIndex" json:"upstream_id"`

	Username    string `json:"username" gorm:"size:24;uniqueIndex" validate:"required,alphanumspace,usernamechangetime"`
	DisplayName string `json:"display_name" gorm:"size:24" validate:"required,alphanumspace,displaynamechangetime"`

	LastUsernameChange    time.Time `json:"last_username_change"`
	LastDisplayNameChange time.Time `json:"last_display_name_change"`

	AuthCodes []AuthCode `gorm:"foreignKey:UserID" json:"-"`
	Caches    []Cache    `gorm:"foreignKey:UserID" json:"-"`

	CreatedAt time.Time `json:"created_at"`
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	var err error
	if err = validate.RegisterValidation("alphanumspace", validateAlphaNumSpace); err != nil {
		panic(err)
	}
	if err = validate.RegisterValidation("usernamechangetime", validateUsernameChangeTime); err != nil {
		panic(err)
	}
	if err = validate.RegisterValidation("displaynamechangetime", validateDisplayNameChangeTime); err != nil {
		panic(err)
	}
}

func validateAlphaNumSpace(fl validator.FieldLevel) bool {
	for _, char := range fl.Field().String() {
		if !(char == ' ' || char == '_' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

func validateUsernameChangeTime(fl validator.FieldLevel) bool {
	lastChange := fl.Field().Interface().(time.Time)
	return time.Since(lastChange).Hours() >= 30*24
}

func validateDisplayNameChangeTime(fl validator.FieldLevel) bool {
	lastChange := fl.Field().Interface().(time.Time)
	return time.Since(lastChange).Hours() >= 7*24
}
