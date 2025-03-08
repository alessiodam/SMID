package main

import (
	_ "embed"
	"github.com/alessiodam/SMID/middleware"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
)

//go:embed index.html
var indexHTML string

//go:embed example.html
var exampleHTML string

//go:embed smid-client.js
var SMIDClientJS string

//go:embed styles.css
var stylesCSS string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed successfully")
		}
	}()

	app := fiber.New()

	app.Use(earlydata.New())

	app.Use(gofiberRecover.New(gofiberRecover.Config{
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			if err := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			}); err != nil {
				log.Printf("Error sending error response: %v", err)
			}
		},
	}))

	app.Use(etag.New())

	app.Use(helmet.New(helmet.Config{
		CrossOriginResourcePolicy: "cross-origin",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: false,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Second,
	}))

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe:     func(c *fiber.Ctx) bool { return true },
		LivenessEndpoint:  "/live",
		ReadinessEndpoint: "/ready",
	}))

	app.Use(gofiberLogger.New(gofiberLogger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(indexHTML)
	})

	app.Get("/dashboard", middleware.AuthMiddleware, handlers.DashboardGetHandler)
	app.Patch("/dashboard", middleware.AuthMiddleware, handlers.DashboardPatchHandler)

	app.Get("/example.html", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(exampleHTML)
	})

	app.Get("/smid-client.js", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/javascript")
		return c.SendString(SMIDClientJS)
	})

	app.Get("/styles.css", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/css")
		return c.SendString(stylesCSS)
	})

	app.Get("/v1/auth-code", handlers.AuthCodeHandler)
	app.Get("/v1/user-id", handlers.UserIdHandler)

	serverErrors := make(chan error, 1)

	go func() {
		log.Println("Starting server on :8000")
		serverErrors <- app.Listen(":8000")
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-signalChan:
		log.Printf("Received signal: %v. Initiating shutdown...", sig)
	}

	shutdownTimeout := 5 * time.Second
	shutdownDone := make(chan struct{})
	go func() {
		if err := app.Shutdown(); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		log.Println("Server shutdown completed gracefully")
	case <-time.After(shutdownTimeout):
		log.Println("Server shutdown timed out")
	}
}
