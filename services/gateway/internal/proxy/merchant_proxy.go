package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

var merchantHTTPClient = &http.Client{Timeout: 10 * time.Second}

// MerchantProxy proxies requests to the merchant service using a direct address,
// following the same pattern as AuthProxy.
func MerchantProxy(merchantAddr, targetPath string) fiber.Handler {
	baseURL := withScheme(merchantAddr)
	return func(c fiber.Ctx) error {
		upstreamPath := targetPath
		for _, p := range c.Route().Params {
			upstreamPath = strings.ReplaceAll(upstreamPath, ":"+p, c.Params(p))
		}

		if len(c.Request().URI().QueryString()) > 0 {
			upstreamPath += "?" + string(c.Request().URI().QueryString())
		}

		body := c.Body()

		req, err := http.NewRequestWithContext(c.Context(), c.Method(), baseURL+upstreamPath, bytes.NewReader(body))
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
		if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
			req.Header.Set("X-User-Id", uid)
		}
		if role, ok := c.Locals("role").(string); ok && role != "" {
			req.Header.Set("X-User-Role", role)
		}

		resp, err := merchantHTTPClient.Do(req)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"error":   fmt.Sprintf("merchant-svc unreachable: %v", err),
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

		c.Set("Content-Type", resp.Header.Get("Content-Type"))
		return c.Status(resp.StatusCode).Send(respBody)
	}
}
