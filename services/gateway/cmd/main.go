package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/moneymate-2026/moneymate-backend/gateway/config"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/middlewares"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/transport/http"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/tracing"
	ws "github.com/moneymate-2026/moneymate-backend/gateway/internal/websocket"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.Tracing.Enabled {
		shutdown, err := tracing.InitTracer(context.Background(), "api-gateway", cfg.Tracing.CollectorURL)
		if err != nil {
			log.Printf("[warn] failed to init tracing: %v (continuing without tracing)", err)
		} else {
			defer shutdown(context.Background())
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	log.Println("[startup] Redis connected")

	authClient, err := proxy.NewAuthClient(cfg.Services.AuthAddr)
	if err != nil {
		log.Fatalf("failed to connect to auth-svc: %v", err)
	}
	defer authClient.Close()
	log.Printf("[startup] auth-svc connected at %s", cfg.Services.AuthAddr)

	hub := ws.NewHub(rdb)
	hub.StartCleanupRoutine(5 * time.Minute)

	serviceRegistry := proxy.NewServiceRegistry(cfg.Services.Downstream)

	authMiddleware := middlewares.RequireAuth(authClient)
	rateLimitMiddleware := middlewares.RateLimiter(
		rdb,
		cfg.RateLimiting.MaxRequests,
		time.Duration(cfg.RateLimiting.WindowSeconds)*time.Second,
	)

	app := fiber.New(fiber.Config{
		AppName:      "MoneyMate API Gateway",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorHandler: http.GlobalErrorHandler,
	})

	app.Use(middlewares.RequestID)
	app.Use(middlewares.Logger)
	app.Use(rateLimitMiddleware)

	http.RegisterRoutes(app, authMiddleware, authClient, serviceRegistry, hub)

	go func() {
		addr := ":" + cfg.Server.Port
		log.Printf("[startup] gateway listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[shutdown] gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(ctx)
	log.Println("[shutdown] gateway stopped")
}
