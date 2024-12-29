package mid

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper format")
)

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(a *auth.Auth, rule string) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, r *http.Request) web.Encoder {
			claims := auth.GetClaims(ctx)
			if claims.Subject == "" {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, no claims")
			}

			var userID uuid.UUID
			id := web.Param(r, "user_id")
			if id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}
				ctx = auth.SetUserID(ctx, userID)
			}

			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

func AuthorizeBusiness(log *logger.Logger, ath *auth.Auth, bsnCore *business.Core) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, r *http.Request) web.Encoder {
			var userID uuid.UUID
			id := web.Param(r, "business_id")

			if id != "" {
				bsnID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				bsn, err := bsnCore.QueryByID(ctx, bsnID)
				if err != nil {
					if errors.Is(err, business.ErrNotFound) {
						return errs.New(errs.Unauthenticated, err)
					}

					return errs.Newf(errs.Internal, "querybyid: bsnID[%s]: %s", bsnID, err)
				}

				userID = bsn.OwnerID
				ctx = setBusiness(ctx, bsn)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

func AuthorizeAppointment(log *logger.Logger, ath *auth.Auth, aptCore *appointment.Core) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, r *http.Request) web.Encoder {
			var userID uuid.UUID
			id := web.Param(r, "appointment_id")

			if id != "" {
				aptID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				apt, err := aptCore.QueryByID(ctx, aptID)
				if err != nil {
					if errors.Is(err, appointment.ErrNotFound) {
						return errs.New(errs.Unauthenticated, err)
					}

					return errs.Newf(errs.Internal, "querybyid: aptID[%s]: %s", aptID, err)
				}

				userID = apt.UserID
				ctx = setAppointment(ctx, apt)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

func AuthorizeGeneralAgenda(log *logger.Logger, ath *auth.Auth, agdCore *agenda.Core, bsnCore *business.Core) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, r *http.Request) web.Encoder {
			var userID uuid.UUID
			id := web.Param(r, "agenda_id")

			if id != "" {
				agdID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				agd, err := agdCore.QueryGeneralAgendaByID(ctx, agdID)
				if err != nil {
					if errors.Is(err, agenda.ErrNotFound) {
						return errs.New(errs.Unauthenticated, err)
					}

					return errs.Newf(errs.Internal, "querybyid: agdID[%s]: %s", agdID, err)
				}
				bsn, err := bsnCore.QueryByID(ctx, agd.BusinessID)
				if err != nil {
					if errors.Is(err, business.ErrNotFound) {
						return errs.New(errs.Unauthenticated, err)
					}

					return errs.Newf(errs.Internal, "querybyid: bsnID[%s]: %s", agd.BusinessID, err)
				}

				userID = bsn.OwnerID
				ctx = setGeneralAgenda(ctx, agd)
				ctx = setBusiness(ctx, bsn)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return errs.Newf(errs.Internal, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

func AuthorizeDailyAgenda(log *logger.Logger, ath *auth.Auth, agdCore *agenda.Core, bsnCore *business.Core) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, r *http.Request) web.Encoder {
			var userID uuid.UUID
			id := web.Param(r, "agenda_id")

			if id != "" {
				agdID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				agd, err := agdCore.QueryDailyAgendaByID(ctx, agdID)
				if err != nil {
					if errors.Is(err, agenda.ErrNotFound) {
						return errs.New(errs.Unauthenticated, err)
					}

					return errs.Newf(errs.Internal, "querybyid: agdID[%s]: %s", agdID, err)
				}
				bsn, err := bsnCore.QueryByID(ctx, agd.BusinessID)
				if err != nil {
					if errors.Is(err, business.ErrNotFound) {
						return errs.Newf(errs.Unauthenticated, "you are not a business owner: %s", err)
					}

					return errs.Newf(errs.Internal, "querybyid: bsnID[%s]: %s", agd.BusinessID, err)
				}

				userID = bsn.OwnerID
				ctx = setDailyAgenda(ctx, agd)
				ctx = setBusiness(ctx, bsn)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}
