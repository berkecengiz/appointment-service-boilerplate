package migrations

import (
	"context"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// Migrations holds the Bun migration registry for the service.
var Migrations = migrate.NewMigrations()

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewCreateTable().
				Model((*models.Appointment)(nil)).
				IfNotExists().
				WithForeignKeys().
				Exec(ctx)
			return err
		},
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*models.Appointment)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	)
}
