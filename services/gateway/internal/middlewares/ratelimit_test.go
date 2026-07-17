package middlewares_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/middlewares"
	"github.com/redis/go-redis/v9"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	app := fiber.New()
	app.Use(middlewares.RateLimiter(rdb, 5, 60*time.Second))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}
}

func TestRateLimiter_RejectsOverLimit(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	app := fiber.New()
	app.Use(middlewares.RateLimiter(rdb, 2, 60*time.Second))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		if i >= 2 && resp.StatusCode != fiber.StatusTooManyRequests {
			t.Errorf("request %d: expected 429, got %d", i, resp.StatusCode)
		}
	}
}
