package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgres creates a Postgres connection pool using configuration values and validates the connection.
func NewPostgres(cfg config.Config) (*sql.DB, error) {
	user := url.QueryEscape(cfg.PGUser)
	pass := url.QueryEscape(cfg.PGPassword)
	host := cfg.PGHost
	port := cfg.PGPort
	dbname := cfg.PGDB
	sslmode := cfg.PGSSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, port, dbname, sslmode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
