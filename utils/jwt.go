package utils

import (
	"errors"
	"github.com/alessiodam/SMID/db"
	"github.com/alessiodam/SMID/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func CreateJWTToken(userID int64) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET environment variable not set")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 12).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GetUserByJWT(tokenString string) (*models.User, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	return nil, errors.New("invalid token")
}
