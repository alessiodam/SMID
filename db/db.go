package db

import (
	"fmt"
	"github.com/alessiodam/SMID/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"os"
	"time"
)

var DB *gorm.DB

type Config struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func DefaultConfig() Config {
	return Config{
		DSN:             os.Getenv("DB_DSN"),
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}
}

func InitDB() error {
	cfg := DefaultConfig()

	if cfg.DSN == "" {
		return fmt.Errorf("database connection string (DSN) is empty")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: gormLogger.Discard,
	})
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("error getting underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := DB.AutoMigrate(&models.User{}, &models.AuthCode{}, &models.Cache{}); err != nil {
		return fmt.Errorf("error running database migrations: %w", err)
	}

	return nil
}

func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("error getting underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}
