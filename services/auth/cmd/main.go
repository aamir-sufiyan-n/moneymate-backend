package main

import (
	"context"
	"log"

	"github.com/moneymate-2026/moneymate-backend/auth/config"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/adapter/postgres"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/adapter/postgres/repo"
	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := postgres.ConnectDB(context.Background(), &postgres.Config{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MinOpenConns:    cfg.Database.MinOpenConns,
		MaxConnLifetime: cfg.Database.MaxConnLifetime,
		MaxIdleTime:     cfg.Database.MaxIdleTime,
	})
	if err != nil {
		log.Fatal(err)
	}

	queries := db.New(pool)
	userRepo := repo.NewUserRepository(queries)

	_ = userRepo

	// continue wiring services...
}