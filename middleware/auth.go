package middleware

import (
	"fmt"
	"strings"

	"github.com/alessiodam/SMID/utils"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		token = c.Cookies("token")
		fmt.Println("cookie token:", token)
		if token == "" {
			return c.Redirect("/")
		}
	} else {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	user, err := utils.GetUserByJWT(token)
	if err != nil {
		return c.Redirect("/")
	}

	c.Locals("user", user)
	return c.Next()
}
