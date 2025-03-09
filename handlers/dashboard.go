package handlers

import (
	"github.com/alessiodam/SMID/db"
	"github.com/alessiodam/SMID/models"
	"github.com/gofiber/fiber/v2"
	"time"
)

func DashboardGetHandler(c *fiber.Ctx) error {
	user := c.Locals("user")
	return c.Render("dashboard.html", fiber.Map{
		"Title": "Dashboard",
		"User":  user,
	})
}

func DashboardPatchHandler(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var payload struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if payload.Username != "" {
		user.Username = payload.Username
		user.LastUsernameChange = time.Now()
	}

	if payload.DisplayName != "" {
		user.DisplayName = payload.DisplayName
		user.LastDisplayNameChange = time.Now()
	}

	if err := db.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
