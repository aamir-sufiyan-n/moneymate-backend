package http

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func GlobalErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	requestID, _ := c.Locals("request_id").(string)

	log.Printf("[error] %s %s → %d: %s", c.Method(), c.Path(), code, message)

	return c.Status(code).JSON(fiber.Map{
		"error":      message,
		"request_id": requestID,
	})
}
