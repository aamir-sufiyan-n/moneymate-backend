package middlewares

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, maxRequests int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		if userID == "" {
			userID = c.IP()
		}
		key := fmt.Sprintf("rate:%s:%s", userID, c.Route().Path)

		count, err := rdb.Incr(c.Context(), key).Result()
		if err != nil {
			return c.Next()
		}

		if count == 1 {
			rdb.Expire(c.Context(), key, window)
		}

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(maxRequests)-count)))

		if count > int64(maxRequests) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
		}

		return c.Next()
	}
}
