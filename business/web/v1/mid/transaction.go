package mid

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
)

func ExecuteInTransaction(log *logger.Logger, bgn transaction.Beginner) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			hasCommited := false

			log.Info(ctx, "BEGIN TRANSACTOIN")
			tx, err := bgn.Begin()
			if err != nil {
				return errs.Newf(errs.Internal, "BEGIN TRANSACTION: %s", err)
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

			ctx = transaction.Set(ctx, tx)

			resp := next(ctx, r)
			if isError(resp) != nil {
				return resp
			}

			log.Info(ctx, "COMMIT TRANSACTION")
			if err := tx.Commit(); err != nil {
				return errs.Newf(errs.Internal, "COMMIT TRANSACTION: %s", err)
			}

			hasCommited = true

			return resp
		}

		return h
	}

	return m
}
