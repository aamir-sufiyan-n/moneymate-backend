package postgres

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)
func RunMigrations(dsn string, migrationsPath string) error {
	log.Println("Running database migrations...")

	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Printf("Migrations path %s not found, using fallback ./migrations", migrationsPath)
		migrationsPath = "./migrations"
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Migrations are already up to date ✅")
			return nil
		}
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	log.Println("Migrations applied successfully ✅")
	return nil
}