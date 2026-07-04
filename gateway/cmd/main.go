package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	db "github.com/moneymate-2026/moneymate-backend/database"
	"github.com/moneymate-2026/moneymate-backend/shared/config"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Fatal error loading config: %v", err)
	}
	pool, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("❌ Fatal error connecting to database: %v", err)
	}
	defer pool.Close()
	db.Pool=pool
	app := fiber.New(fiber.Config{
		AppName: "MoneyMate Service",
	})


	app.Get("/health", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		dbStatus := "up"
		if err := db.Pool.Ping(ctx); err != nil {
			dbStatus = "down"
			log.Printf("⚠️ Healthcheck DB Ping Failed: %v", err)
		}

		overallStatus := "healthy"
		statusCode := fiber.StatusOK

		if dbStatus == "down" {
			overallStatus = "unhealthy"
			statusCode = fiber.StatusServiceUnavailable // 503 error
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"message":"get the flooooo offf",
			"status":    overallStatus,
			"timestamp": time.Now().Format(time.RFC3339),
			"services": fiber.Map{
				"api":      "up",
				"database": dbStatus,
			},
		})
	})

	port := cfg.Gateway.HTTPAddr
	if port == "" {
		port = ":8080"
	}

	log.Printf("🚀 Server starting on %s...", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}


