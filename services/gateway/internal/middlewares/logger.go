package middlewares

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logger(c fiber.Ctx) error {
	start := time.Now()

	err := c.Next()

	latency := time.Since(start)
	requestID, _ := c.Locals("request_id").(string)

	log.Printf("[%s] %s %s → %d (%v)",
		requestID,
		c.Method(),
		c.Path(),
		c.Response().StatusCode(),
		latency,
	)

	return err
}
