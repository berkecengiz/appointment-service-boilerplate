package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// NewPostgres creates a Postgres Bun database using configuration values and validates the connection.
func NewPostgres(cfg config.Config) (*bun.DB, error) {
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

	connector := pgdriver.NewConnector(pgdriver.WithDSN(dsn))

	sqldb := sql.OpenDB(connector)

	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxIdleTime(5 * time.Minute)
	sqldb.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqldb.PingContext(ctx); err != nil {
		_ = sqldb.Close()
		return nil, err
	}

	db := bun.NewDB(sqldb, pgdialect.New())
	return db, nil
}
