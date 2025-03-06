package handlers

import (
	"errors"
	"github.com/alessiodam/SMID/db"
	"github.com/alessiodam/SMID/models"
	"github.com/alessiodam/SMID/utils"
	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"time"
)

var snowflakeNode *snowflake.Node

func init() {
	var err error
	snowflakeNode, err = snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("Failed to create snowflake node: %v", err)
	}
}

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

	if userID == 0 {
		source = "upstream"
		upstreamID, err := utils.PerformUpstreamRequest(domain, phpSessId)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		var user models.User
		result := db.DB.Where("upstream_id = ?", upstreamID).First(&user)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = models.User{
				ID:         snowflakeNode.Generate().Int64(),
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
	var authCode models.AuthCode
	if err := db.DB.Where("code = ?", code).First(&authCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Code not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if time.Since(authCode.CreatedAt) > time.Minute {
		db.DB.Delete(&authCode)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Code expired"})
	}
	return c.JSON(fiber.Map{"user_id": authCode.UserID})
}

func getCachedUser(phpHash string) (int64, error) {
	var cache models.Cache
	result := db.DB.Where("php_hash = ?", phpHash).First(&cache)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if result.Error != nil {
		return 0, result.Error
	}
	if time.Now().Unix() > cache.ExpiresAt {
		db.DB.Delete(&cache)
		return 0, nil
	}
	return cache.UserID, nil
}

func cacheUser(phpHash string, userID int64) error {
	cache := models.Cache{
		PhpHash:   phpHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(2 * 24 * time.Hour).Unix(),
	}
	return db.DB.Save(&cache).Error
}
