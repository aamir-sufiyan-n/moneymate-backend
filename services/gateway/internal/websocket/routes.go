package websocket

import (
	"github.com/gofiber/fiber/v3"

	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
)

func RegisterWebSocketRoutes(app *fiber.App, hub *Hub, authClient proxy.AuthClient) {
	app.Get("/ws", hub.HandleConnection(authClient))
}
