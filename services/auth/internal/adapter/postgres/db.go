package postgres

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
)

type Database struct {
	DB      *sql.DB
	Queries *db.Queries
}

func New(dbConn *sql.DB) *Database {
	return &Database{
		DB:      dbConn,
		Queries: db.New(dbConn),
	}
}

func Connect(dsn string) (*Database, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(time.Hour)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return New(conn), nil
}
