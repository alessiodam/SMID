package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alessiodam/SMID/db"
	"github.com/alessiodam/SMID/models"
	"github.com/alessiodam/SMID/utils"
	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	validDomainSuffix = ".smartschool.be"
	authCodeTTL       = time.Minute
	cacheTTL          = 2 * 24 * time.Hour
)

var snowflakeNode *snowflake.Node

func init() {
	var err error
	snowflakeNode, err = snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("Failed to create snowflake node: %v", err)
	}
}

func respondError(c *fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(fiber.Map{"error": err.Error()})
}

func AuthCodeHandler(c *fiber.Ctx) error {
	domain := strings.TrimSpace(c.Query("domain"))
	phpSessId := strings.TrimSpace(c.Query("phpSessId"))

	if domain == "" || phpSessId == "" {
		return respondError(c, http.StatusBadRequest, errors.New("missing domain or phpSessId query parameters"))
	}
	if !strings.HasSuffix(domain, validDomainSuffix) {
		return respondError(c, http.StatusBadRequest, errors.New("domain must end with "+validDomainSuffix))
	}

	phpHash := utils.HashPhpSessId(phpSessId)
	userID, err := getCachedUser(phpHash)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, err)
	}

	var source string
	if userID == 0 {
		source = "upstream"
		upstreamID, err := utils.PerformUpstreamRequest(domain, phpSessId)
		if err != nil {
			return respondError(c, http.StatusInternalServerError, err)
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
				return respondError(c, http.StatusInternalServerError, err)
			}
		} else if result.Error != nil {
			return respondError(c, http.StatusInternalServerError, result.Error)
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
		return respondError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(fiber.Map{"code": code, "source": source})
}

func UserIdHandler(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		return respondError(c, http.StatusBadRequest, errors.New("missing code query parameter"))
	}

	var authCode models.AuthCode
	if err := db.DB.Where("code = ?", code).First(&authCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return respondError(c, http.StatusNotFound, errors.New("code not found"))
		}
		return respondError(c, http.StatusInternalServerError, err)
	}

	if time.Since(authCode.CreatedAt) > authCodeTTL {
		db.DB.Delete(&authCode)
		return respondError(c, http.StatusUnauthorized, errors.New("code expired"))
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
	if time.Since(cache.CreatedAt) > cacheTTL {
		db.DB.Delete(&cache)
		return 0, nil
	}
	return cache.UserID, nil
}

func cacheUser(phpHash string, userID int64) error {
	cache := models.Cache{
		PhpHash:   phpHash,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	return db.DB.Save(&cache).Error
}
