package migrations

import (
	"context"
	"strings"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.NewCreateTable().Model((*models.Client)(nil)).IfNotExists().WithForeignKeys().Exec(ctx); err != nil {
				return err
			}

			if _, err := db.NewCreateTable().Model((*models.Provider)(nil)).IfNotExists().WithForeignKeys().Exec(ctx); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE appointments RENAME COLUMN customerid TO clientid`); err != nil {
				if !strings.Contains(err.Error(), "does not exist") {
					return err
				}
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE appointments ADD CONSTRAINT appointments_clientid_fkey FOREIGN KEY (clientid) REFERENCES clients(id)`); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return err
				}
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE appointments ADD CONSTRAINT appointments_providerid_fkey FOREIGN KEY (providerid) REFERENCES providers(id)`); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return err
				}
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `ALTER TABLE appointments DROP CONSTRAINT IF EXISTS appointments_providerid_fkey`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE appointments DROP CONSTRAINT IF EXISTS appointments_clientid_fkey`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `ALTER TABLE appointments RENAME COLUMN clientid TO customerid`); err != nil {
				if !strings.Contains(err.Error(), "does not exist") {
					return err
				}
			}

			if _, err := db.NewDropTable().Model((*models.Provider)(nil)).IfExists().Exec(ctx); err != nil {
				return err
			}

			if _, err := db.NewDropTable().Model((*models.Client)(nil)).IfExists().Exec(ctx); err != nil {
				return err
			}

			return nil
		},
	)
}
