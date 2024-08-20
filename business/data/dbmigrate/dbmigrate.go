package dbmigrate

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	database "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ardanlabs/darwin/v3"
	"github.com/ardanlabs/darwin/v3/dialects/postgres"
	"github.com/ardanlabs/darwin/v3/drivers/generic"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/migrate.sql
	migrateDoc string

	//go:embed sql/seed.sql
	seedDoc string
)

func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	driver, err := generic.New(db.DB, postgres.Dialect{})
	if err != nil {
		return fmt.Errorf("construct darwin driver: %w", err)
	}

	d := darwin.New(driver, darwin.ParseMigrations(migrateDoc))
	d.Migrate()

	return nil
}

func Seed(ctx context.Context, db *sqlx.DB) (err error) {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if errTx := tx.Rollback(); errTx != nil {
			if errors.Is(errTx, sql.ErrTxDone) {
				return
			}
			err = fmt.Errorf("rollback: %w", err)
			return
		}
	}()

	if _, err := tx.Exec(seedDoc); err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
