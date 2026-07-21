package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
)

var authHTTPClient = &http.Client{Timeout: 10 * time.Second}

// AuthProxy returns a Fiber handler that transparently proxies requests
// to the given auth-svc path. The request body, Content-Type, and
// important headers (X-Device-Id, User-Agent, Authorization) are forwarded.
// The auth-svc response (status + JSON body) is returned as-is.
func AuthProxy(authAddr, targetPath string) fiber.Handler {
	baseURL := "http://" + authAddr
	return func(c fiber.Ctx) error {
		body := c.Body()

		req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, baseURL+targetPath, bytes.NewReader(body))
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"success": false,
				"error":   "failed to create upstream request",
			})
		}

		req.Header.Set("Content-Type", "application/json")
		if v := c.Get("X-Device-Id"); v != "" {
			req.Header.Set("X-Device-Id", v)
		}
		if v := c.Get("Authorization"); v != "" {
			req.Header.Set("Authorization", v)
		}
		if v := c.Get("User-Agent"); v != "" {
			req.Header.Set("User-Agent", v)
		}

		resp, err := authHTTPClient.Do(req)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"error":   fmt.Sprintf("auth-svc unreachable: %v", err),
			})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"success": false,
				"error":   "failed to read upstream response",
			})
		}

		return c.Status(resp.StatusCode).Send(respBody)
	}
}

// AuthProxyGET returns a Fiber handler that proxies GET requests to auth-svc.
func AuthProxyGET(authAddr, targetPath string) fiber.Handler {
	baseURL := "http://" + authAddr
	return func(c fiber.Ctx) error {
		req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, baseURL+targetPath, nil)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"success": false,
				"error":   "failed to create upstream request",
			})
		}

		req.Header.Set("Content-Type", "application/json")
		if v := c.Get("Authorization"); v != "" {
			req.Header.Set("Authorization", v)
		}

		resp, err := authHTTPClient.Do(req)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"error":   fmt.Sprintf("auth-svc unreachable: %v", err),
			})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"success": false,
				"error":   "failed to read upstream response",
			})
		}

		return c.Status(resp.StatusCode).Send(respBody)
	}
}
