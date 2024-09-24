package dbmigrate

import (
	"context"
	"database/sql"
	"embed"
	_ "embed"
	"errors"
	"fmt"

	database "github.com/ameghdadian/service/business/data/dbsql/pgx"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

//go:embed sql/migrations/*.sql
var migrateDoc embed.FS

//go:embed sql/seed.sql
var seedDoc string

func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	d, err := iofs.New(migrateDoc, "sql/migrations")
	if err != nil {
		return fmt.Errorf("loading migration files into migrate iofs: %w", err)
	}

	instance, err := pgx.WithInstance(db.DB, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("creating pgx driver instance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", instance)
	if err != nil {
		return fmt.Errorf("constructing migrate driver: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrating the database: %w", err)
	}

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
