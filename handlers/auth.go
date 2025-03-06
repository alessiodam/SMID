package handlers

import (
	"errors"
	"github.com/alessiodam/SMID/db"
	"github.com/alessiodam/SMID/models"
	"github.com/alessiodam/SMID/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AuthCodeHandler(c *fiber.Ctx) error {
	domain := c.Query("domain")
	phpSessId := c.Query("phpSessId")
	if domain == "" || phpSessId == "" || !strings.HasSuffix(domain, "smartschool.be") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Missing domain or phpSessId query parameters, or domain does not end with smartschool.be"})
	}

	phpHash := utils.HashPhpSessId(phpSessId)
	userID, err := getCachedUser(phpHash)
	var source string
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if userID == "" {
		source = "upstream"
		upstreamID, err := utils.PerformUpstreamRequest(domain, phpSessId)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		var user models.User
		result := db.DB.Where("upstream_id = ?", upstreamID).First(&user)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = models.User{
				ID:         uuid.New().String(),
				UpstreamID: upstreamID,
				CreatedAt:  time.Now(),
			}
			if err := db.DB.Create(&user).Error; err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
		} else if result.Error != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": result.Error.Error()})
		}
		userID = user.ID
		if err := cacheUser(phpHash, userID); err != nil {
			log.Printf("Error caching user: %v", err)
		}
	} else {
		source = "cache"
	}

	code := uuid.New().String()
	auth := models.AuthCode{
		Code:      code,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	if err := db.DB.Create(&auth).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"code": code, "source": source})
}

func UserIdHandler(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Missing code query parameter"})
	}
	var auth models.AuthCode
	if err := db.DB.Where("code = ?", code).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Code not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"user_id": auth.UserID})
}

func getCachedUser(phpHash string) (string, error) {
	var cache models.Cache
	result := db.DB.Where("php_hash = ?", phpHash).First(&cache)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if result.Error != nil {
		return "", result.Error
	}
	if time.Now().Unix() > cache.ExpiresAt {
		db.DB.Delete(&cache)
		return "", nil
	}
	return cache.UserID, nil
}

func cacheUser(phpHash, userID string) error {
	cache := models.Cache{
		PhpHash:   phpHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(2 * 24 * time.Hour).Unix(),
	}
	return db.DB.Save(&cache).Error
}
