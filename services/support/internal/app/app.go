package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/abijith/moneymate-backend/services/support/config"
	"github.com/abijith/moneymate-backend/services/support/internal/repository/postgres"
	"github.com/abijith/moneymate-backend/services/support/internal/usecase"
	transporthttp "github.com/abijith/moneymate-backend/services/support/internal/transport/http"
)

type App struct {
	HTTPServer  *fiber.App
	DB          *sql.DB
	RDB         *redis.Client
	Config      *config.Config
	HTTPAddr    string
	ChatUseCase usecase.ChatUseCase
}

func Build(cfg *config.Config) (*App, error) {
	dbURL := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=support",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SslMode,
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	migrationDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=support",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SslMode,
	)
	if err := postgres.RunMigrations(migrationDSN, cfg.Database.MigrationsPath); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	supportRepo := postgres.NewSupportRepo(db)
	supportUseCase := usecase.NewSupportUseCase(supportRepo)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})

	chatRepo := postgres.NewChatRepo(db)
	chatUseCase := usecase.NewChatUseCase(chatRepo, rdb)

	supportHandler := transporthttp.NewSupportHandler(supportUseCase)
	chatHandler := transporthttp.NewChatHandler(chatUseCase)
	httpServer := setupHTTPServer(supportHandler, chatHandler)

	httpAddr := cfg.Server.HTTPAddr
	if port := os.Getenv("PORT"); port != "" {
		httpAddr = port
	}
	if httpAddr == "" {
		httpAddr = "8085"
	}
	if !strings.Contains(httpAddr, ":") {
		httpAddr = "0.0.0.0:" + httpAddr
	}

	return &App{
		HTTPServer:  httpServer,
		DB:          db,
		RDB:         rdb,
		Config:      cfg,
		HTTPAddr:    httpAddr,
		ChatUseCase: chatUseCase,
	}, nil
}

func setupHTTPServer(supportHandler *transporthttp.SupportHandler, chatHandler *transporthttp.ChatHandler) *fiber.App {
	server := fiber.New(fiber.Config{
		AppName: "support-service",
	})

	server.Use(recover.New())
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	transporthttp.RegisterRoutes(server, supportHandler, chatHandler)
	transporthttp.RegisterAdminRoutes(server, supportHandler, chatHandler)

	return server
}

func (a *App) Run() error {
	go a.StartRedisSubscriber()
	log.Printf("Starting HTTP server on %s", a.HTTPAddr)
	return a.HTTPServer.Listen(a.HTTPAddr)
}

func (a *App) StartRedisSubscriber() {
	pubsub := a.RDB.Subscribe(context.Background(), "incoming_ws_messages")
	defer pubsub.Close()
	ch := pubsub.Channel()
	for msg := range ch {
		var p struct {
			SenderID     string `json:"sender_id"`
			SenderType   string `json:"sender_type"`
			ReceiverID   string `json:"receiver_id"`
			ReceiverType string `json:"receiver_type"`
			Message      string `json:"message"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &p); err == nil {
			sID, _ := uuid.Parse(p.SenderID)
			rID, _ := uuid.Parse(p.ReceiverID)
			_, err := a.ChatUseCase.SendMessage(context.Background(), sID, p.SenderType, rID, p.ReceiverType, p.Message)
			if err != nil {
				log.Printf("failed to send websocket chat msg: %v", err)
			}
		}
	}
}

func (a *App) Close() {
	if a.HTTPServer != nil {
		a.HTTPServer.Shutdown()
	}
	if a.DB != nil {
		a.DB.Close()
	}
	if a.RDB != nil {
		a.RDB.Close()
	}
}
