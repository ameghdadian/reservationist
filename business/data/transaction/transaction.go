package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ameghdadian/service/foundation/logger"
)

type Transaction interface {
	Commit() error
	Rollback() error
}

type Beginner interface {
	Begin() (Transaction, error)
}

// ========================================================

type ctxKey int

const trKey ctxKey = 2

func Set(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, trKey, tx)
}

func Get(ctx context.Context) (Transaction, bool) {
	v, ok := ctx.Value(trKey).(Transaction)
	return v, ok
}

// ========================================================

// ExecuteUnderTransaction should ONLY be used when writing tests to run a handler under transaction conditions.
func ExecuteUnderTransaction(ctx context.Context, log *logger.Logger, bgn Beginner, fn func(tx Transaction) error) error {
	hasCommited := false

	log.Info(ctx, "BEGIN TRANSACTION")
	tx, err := bgn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if !hasCommited {
			log.Info(ctx, "ROLLBACK TRANSACTION")
		}

		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			log.Info(ctx, "ROLLBACK TRANSACTION", "ERROR", err)
		}
	}()

	if err := fn(tx); err != nil {
		return fmt.Errorf("EXECUTE TRANSACTION: %w", err)
	}

	log.Info(ctx, "COMMIT TRANSACTION")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("COMMIT TRANSACTION: %w", err)
	}

	hasCommited = true

	return nil
}
