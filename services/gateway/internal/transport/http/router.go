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
	authAddr string,
) {
	api := app.Group("/api/v1")

	api.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"service": "gateway",
		})
	})

	userAuth := api.Group("/auth")
	userAuth.Post("/register", proxy.AuthProxy(authAddr, "/user/auth/register"))
	userAuth.Post("/login", proxy.AuthProxy(authAddr, "/user/auth/login"))
	userAuth.Post("/otp/send", proxy.AuthProxy(authAddr, "/user/auth/otp/send"))
	userAuth.Post("/otp/verify", proxy.AuthProxy(authAddr, "/user/auth/otp/verify"))

	merchantAuth := api.Group("/merchant/auth")
	merchantAuth.Post("/register", proxy.AuthProxy(authAddr, "/merchant/auth/register"))
	merchantAuth.Post("/login", proxy.AuthProxy(authAddr, "/merchant/auth/login"))
	merchantAuth.Post("/otp/send", proxy.AuthProxy(authAddr, "/merchant/auth/otp/send"))
	merchantAuth.Post("/otp/verify", proxy.AuthProxy(authAddr, "/merchant/auth/otp/verify"))

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
