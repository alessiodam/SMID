package db

import (
	"github.com/alessiodam/SMID/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{
		Logger: gormLogger.Discard,
	})
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	if err := DB.AutoMigrate(&models.User{}, &models.AuthCode{}, &models.Cache{}); err != nil {
		log.Fatalf("Error migrating DB: %v", err)
	}
}
