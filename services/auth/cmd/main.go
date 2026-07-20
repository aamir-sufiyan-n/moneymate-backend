package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/moneymate-2026/moneymate-backend/auth/config"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/adapter/postgres"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/adapter/postgres/repo"
	rediscard "github.com/moneymate-2026/moneymate-backend/auth/internal/adapter/redis"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/infra/hasher"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/infra/idgen"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/infra/mailer"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/infra/tokenissuer"
	transporthttp "github.com/moneymate-2026/moneymate-backend/auth/internal/transport/http"
	usecase "github.com/moneymate-2026/moneymate-backend/auth/internal/usecases"
	sharedjwt "github.com/moneymate-2026/moneymate-backend/shared/pkg/jwt"
	sharedmailer "github.com/moneymate-2026/moneymate-backend/shared/pkg/mailer"
	sharedpgxtx "github.com/moneymate-2026/moneymate-backend/shared/pkg/pgxtx"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := postgres.ConnectDB(context.Background(), &postgres.Config{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MinOpenConns:    cfg.Database.MinOpenConns,
		MaxConnLifetime: cfg.Database.MaxConnLifetime,
		MaxIdleTime:     cfg.Database.MaxIdleTime,
	})
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	redisClient, err := rediscard.NewClient(rediscard.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       0,
	})
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisClient.Close()

	userRepo := repo.NewUserRepo(pool)
	roleRepo := repo.NewRoleRepo(pool)
	refreshTokenRepo := repo.NewRefreshTokenRepo(pool)
	store := rediscard.NewStore(redisClient)
	txMgr := sharedpgxtx.New(pool)

	h := hasher.New()
	g := idgen.New()
	issuer := tokenissuer.New(sharedjwt.Config{
		AccessSecret:     cfg.JWT.AccessSecret,
		RefreshSecret:    cfg.JWT.RefreshSecret,
		AccessExpiryMins: cfg.JWT.AccessExpiryMinutes,
		RefreshExpiryHrs: cfg.JWT.RefreshExpiryHours,
	})

	mailerClient := sharedmailer.New(sharedmailer.Config{
		Host:        cfg.SMTP.Host,
		Port:        cfg.SMTP.Port,
		Username:    cfg.SMTP.Username,
		Password:    cfg.SMTP.Password,
		FromAddress: cfg.SMTP.FromAddress,
		FromName:    cfg.SMTP.FromName,
	})
	otpMailer := mailer.NewOtpMail(mailerClient)

	authUC := usecase.NewAuthUsecase(userRepo, roleRepo, refreshTokenRepo, store, txMgr, h, g, issuer, sharedjwt.Config{
		AccessSecret:     cfg.JWT.AccessSecret,
		RefreshSecret:    cfg.JWT.RefreshSecret,
		AccessExpiryMins: cfg.JWT.AccessExpiryMinutes,
		RefreshExpiryHrs: cfg.JWT.RefreshExpiryHours,
	})
	otpUC := usecase.NewOTPUsecase(userRepo, store, otpMailer, cfg.OTP)

	authHandler := transporthttp.NewAuthHandler(authUC, otpUC, userRepo, cfg.JWT.AccessSecret)

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		AppName:      "auth-service",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Device-Id"},
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "auth"})
	})

	noopAuth := func(c fiber.Ctx) error { return c.Next() }
	transporthttp.RegisterRoutes(app, authHandler, noopAuth)

	addr := cfg.Server.HTTPAddr
	if addr == "" {
		addr = ":8081"
	}

	go func() {
		log.Printf("auth HTTP server listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down auth service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("auth service stopped")
}
