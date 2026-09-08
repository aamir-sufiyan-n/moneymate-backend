package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
)

// ServiceRegistry maps service names to their HTTP addresses.
type ServiceRegistry struct {
	services map[string]string 
}

func NewServiceRegistry(services map[string]string) *ServiceRegistry {
	return &ServiceRegistry{services: services}
}

// GetAddress returns the HTTP address of the named service.
func (r *ServiceRegistry) GetAddress(serviceName string) (string, error) {
	addr, ok := r.services[serviceName]
	if !ok {
		return "", fmt.Errorf("unknown service: %s", serviceName)
	}
	return addr, nil
}

// ProxyToService creates a Fiber handler that proxies the request to a downstream HTTP service.
func ProxyToService(registry *ServiceRegistry, serviceName string) fiber.Handler {
	return func(c fiber.Ctx) error {
		addr, err := registry.GetAddress(serviceName)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is not available yet", serviceName),
			})
		}


		targetPath := strings.TrimPrefix(c.OriginalURL(), "/api/v1")
		url := "http://" + addr + targetPath
		
		if err := proxy.Do(c, url); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to proxy to %s: %v", serviceName, err),
			})
		}
		
		return nil
	}
}

func ExtractServiceName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}

func HTTPProxy(registry *ServiceRegistry, serviceName string, targetPath string) fiber.Handler {
	return func(c fiber.Ctx) error {
		addr, err := registry.GetAddress(serviceName)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
		}
		
		baseURL := addr
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			baseURL = "http://" + baseURL
		}

		upstreamPath := targetPath
		for _, p := range c.Route().Params {
			upstreamPath = strings.ReplaceAll(upstreamPath, ":"+p, c.Params(p))
		}
		
		if len(c.Request().URI().QueryString()) > 0 {
			upstreamPath += "?" + string(c.Request().URI().QueryString())
		}
		
		req, err := http.NewRequestWithContext(c.Context(), c.Method(), baseURL+upstreamPath, bytes.NewReader(c.Body()))
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to create request"})
		}
		
		c.Request().Header.VisitAll(func(k, v []byte) {
			key := string(k)
			if key != "Host" && key != "Connection" {
				req.Header.Set(key, string(v))
			}
		})
		if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
			req.Header.Set("X-User-Id", uid)
		}
		if role, ok := c.Locals("role").(string); ok && role != "" {
			req.Header.Set("X-User-Role", role)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "upstream unreachable: " + err.Error()})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to read response"})
		}

		for k, v := range resp.Header {
			c.Set(k, v[0])
		}
		return c.Status(resp.StatusCode).Send(respBody)
	}
}