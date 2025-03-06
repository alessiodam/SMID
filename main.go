package main

import (
	_ "embed"
	"github.com/alessiodam/SMID/db"
	"github.com/alessiodam/SMID/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	gofiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	gofiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"log"
	"time"
)

//go:embed index.html
var indexHTML string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	db.InitDB()
	app := fiber.New()
	app.Use(earlydata.New())
	app.Use(gofiberRecover.New(gofiberRecover.Config{
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			err := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unknown error occurred",
			})
			if err != nil {
				log.Println("Error while handling error")
			}
		},
	}))
	app.Use(etag.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: false,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Second,
	}))
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint:  "/live",
		ReadinessEndpoint: "/ready",
	}))
	app.Use(gofiberLogger.New(gofiberLogger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(indexHTML)
	})

	app.Get("/v1/auth-code", handlers.AuthCodeHandler)
	app.Get("/v1/user-id", handlers.UserIdHandler)
	log.Fatal(app.Listen(":8000"))
}
