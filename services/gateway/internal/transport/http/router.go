package http

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/moneymate-2026/moneymate-backend/gateway/internal/middlewares"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
	ws "github.com/moneymate-2026/moneymate-backend/gateway/internal/websocket"
)

func RegisterRoutes(
	app *fiber.App,
	authMiddleware fiber.Handler,
	authClient proxy.AuthClient,
	registry *proxy.ServiceRegistry,
	hub *ws.Hub,
) {
	api := app.Group("/api/v1")

	api.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"service": "gateway",
		})
	})

	authGroup := api.Group("/auth")
	authGroup.Post("/login", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "login pending gRPC contract",
		})
	})
	authGroup.Post("/register", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "register pending gRPC contract",
		})
	})

	secure := api.Group("/secure")
	secure.Use(authMiddleware)
	secure.Get("/profile", func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"user_id": userID,
			},
		})
	})

	merchant := api.Group("/merchant")
	merchant.Use(authMiddleware)
	merchant.Use(middlewares.RequireRole("merchant"))
	merchant.Get("/dashboard", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "merchant dashboard placeholder",
		})
	})

	downstreamServices := []string{"payment", "merchant", "campaign", "debt", "pod", "scheduler", "referral", "rewards", "routing", "notification"}
	for _, svc := range downstreamServices {
		svcName := svc
		api.All(fmt.Sprintf("/%s/*", svcName), authMiddleware, proxy.ProxyToService(registry, svcName))
	}

	ws.RegisterWebSocketRoutes(app, hub, authClient)
}
