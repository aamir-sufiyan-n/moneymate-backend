package main

import (
	"log"

	"github.com/moneymate-2026/moneymate-backend/auth/config"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/adapter/postgres"
)

func main() {
	cfg := config.Load()

	database, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := postgres.NewUserRepository(database.Queries)

	_ = userRepo

	// continue wiring services...
}