package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
)

// RequireAuth creates a Fiber middleware that validates the JWT by calling auth-svc
// via gRPC. On success, it injects user_id, role, and optionally merchant_id into
// the Fiber context so downstream handlers can access them via c.Locals("user_id").
func RequireAuth(authClient proxy.AuthClient) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Extract the Authorization header: "Bearer <token>"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format, expected: Bearer <token>",
			})
		}

		token := parts[1]

		// Call auth-svc gRPC to verify the token
		claims, err := authClient.VerifyAccessToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Inject identity into context — downstream handlers and RBAC middleware read these
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		if claims.MerchantID != "" {
			c.Locals("merchant_id", claims.MerchantID)
		}

		return c.Next()
	}
}

// RequireTransactionAuth is similar but validates a short-lived transaction token.
// Used on payment endpoints above the user's configured threshold (see doc.md §5.8).
func RequireTransactionAuth(authClient proxy.AuthClient) fiber.Handler {
	return func(c fiber.Ctx) error {
		transactionToken := c.Get("X-Transaction-Token")
		if transactionToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing x-transaction-token header",
			})
		}

		transactionID := c.Get("X-Transaction-ID")
		if transactionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "missing x-transaction-id header",
			})
		}

		claims, err := authClient.VerifyTransactionToken(c.Context(), transactionToken, transactionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired transaction token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("transaction_id", claims.TransactionID)

		return c.Next()
	}
}