package middlewares

import (
	"github.com/gofiber/fiber/v3"
)

// RequireRole creates a middleware that checks the role claim injected by RequireAuth.
// Usage: app.Get("/merchant/dashboard", RequireRole("merchant"), handler)
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied: no role found",
			})
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied: insufficient permissions",
		})
	}
}