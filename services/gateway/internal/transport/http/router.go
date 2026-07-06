package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	
	// Update this specific line!
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
)

type Router struct {
	app        *fiber.App
	authClient proxy.AuthClient
}

func NewRouter(auth proxy.AuthClient) *Router {
	// Optimize Fiber for mechanical sympathy
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Prefork:               false, // Enable in prod for multi-core socket listening
		ReduceMemoryUsage:     true,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	return &Router{
		app:        app,
		authClient: auth,
	}
}

func (r *Router) SetupRoutes() {
	api := r.app.Group("/api/v1")

	// Health check for load balancers
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Example route using the abstract AuthClient
	api.Get("/verify-test", func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		
		// Call the interface (currently mocked, later gRPC)
		userID, err := r.authClient.VerifyToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		return c.JSON(fiber.Map{"user_id": userID})
	})
}

func (r *Router) Listen(addr string) error {
	return r.app.Listen(addr)
}

func (r *Router) Shutdown() error {
	return r.app.Shutdown()
}