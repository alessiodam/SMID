package middleware

import (
	"github.com/alessiodam/SMID/utils"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		token = c.Cookies("token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
		}
	} else {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	user, err := utils.GetUserByJWT(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	c.Locals("user", user)
	return c.Next()
}
