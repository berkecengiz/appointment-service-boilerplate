package db

import (
	"context"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/db/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// RunMigrations applies pending migrations using Bun's migrator, ensuring schema is up to date.
func RunMigrations(ctx context.Context, db *bun.DB) error {
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return err
	}

	if err := migrator.Lock(ctx); err != nil {
		return err
	}
	defer func() { _ = migrator.Unlock(ctx) }()

	if _, err := migrator.Migrate(ctx); err != nil {
		return err
	}
	return nil
}
