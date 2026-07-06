package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	// Using YOUR actual module path
	"github.com/moneymate-2026/moneymate-backend/gateway/config"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/transport/http"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Initialize abstract clients (Swap to NewGrpcAuthClient later)
	authClient := proxy.NewMockAuthClient()

	// 2. Initialize router and inject dependencies
	router := http.NewRouter(authClient)
	router.SetupRoutes()

	// 3. Start server in a non-blocking goroutine
	go func() {
		log.Printf("Gateway starting on port %s", cfg.AppPort)
		if err := router.Listen(cfg.AppPort); err != nil {
			log.Fatalf("Gateway server failed: %v", err)
		}
	}()

	// 4. Mechanical sympathy: Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Gracefully shutting down Gateway...")
	if err := router.Shutdown(); err != nil {
		log.Fatalf("Gateway forced to shutdown: %v", err)
	}
	log.Println("Gateway stopped cleanly.")
}