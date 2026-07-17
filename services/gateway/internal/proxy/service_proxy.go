package proxy

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceRegistry maps service names to their gRPC addresses.
// Updated as new services come online.
type ServiceRegistry struct {
	services map[string]string // name → "host:port"
}

// NewServiceRegistry creates a registry from config.
// Example: {"payment": "payment-svc:9092", "merchant": "merchant-svc:9093"}
func NewServiceRegistry(services map[string]string) *ServiceRegistry {
	return &ServiceRegistry{services: services}
}

// GetConn returns a gRPC connection to the named service.
// Connections are NOT pooled here — for production, use a connection pool
// or let gRPC's built-in pooling handle it.
func (r *ServiceRegistry) GetConn(serviceName string) (*grpc.ClientConn, error) {
	addr, ok := r.services[serviceName]
	if !ok {
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}

	return grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// ProxyToService creates a Fiber handler that proxies the request to a downstream
// gRPC service. For now, this is a placeholder that returns 503 for missing services.
// As each service comes online, its handler will be registered directly (not via proxy)
// until we implement full gRPC-gateway transcoding.
func ProxyToService(registry *ServiceRegistry, serviceName string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Until downstream services are built, return a clear unavailable response.
		// Once payment-svc exists, replace this with actual gRPC forwarding.
		_, err := registry.GetConn(serviceName)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is not available yet", serviceName),
			})
		}

		// TODO: Implement full gRPC transcoding here as each service comes online.
		// For now, services will be wired directly in the router.
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fmt.Sprintf("%s proxy not yet implemented", serviceName),
		})
	}
}

// ExtractServiceName parses the route path "/api/v1/payment/..." and returns "payment".
func ExtractServiceName(path string) string {
	// path format: /api/v1/{service}/...
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}