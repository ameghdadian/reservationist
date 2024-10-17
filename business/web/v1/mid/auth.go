package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/response"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper format")
)

func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewAuthError("authenticate: failed %s", err)
			}

			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(a *auth.Auth, rule string) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims := auth.GetClaims(ctx)
			if claims.Subject == "" {
				return auth.NewAuthError("authorize: you are not authorized for that action, no claims")
			}

			var userID uuid.UUID
			id := web.Param(r, "user_id")
			if id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}
				ctx = auth.SetUserID(ctx, userID)
			}

			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func AuthorizeBusiness(log *logger.Logger, ath *auth.Auth, bsnCore *business.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID
			id := web.Param(r, "business_id")

			if id != "" {
				bsnID, err := uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}

				bsn, err := bsnCore.QueryByID(ctx, bsnID)
				if err != nil {
					if errors.Is(err, business.ErrNotFound) {
						return response.NewError(err, http.StatusNotFound)
					}

					return fmt.Errorf("querybyid: bsnID[%s]: %w", bsnID, err)
				}

				userID = bsn.OwnerID
				setBusiness(ctx, bsn)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func AuthorizeAppointment(log *logger.Logger, ath *auth.Auth, aptCore *appointment.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID
			id := web.Param(r, "appointment_id")

			if id != "" {
				aptID, err := uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}

				apt, err := aptCore.QueryByID(ctx, aptID)
				if err != nil {
					if errors.Is(err, appointment.ErrNotFound) {
						return response.NewError(err, http.StatusNotFound)
					}

					return fmt.Errorf("querybyid: aptID[%s]: %w", aptID, err)
				}

				userID = apt.UserID
				setAppointment(ctx, apt)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func AuthorizeGeneralAgenda(log *logger.Logger, ath *auth.Auth, agdCore *agenda.Core, bsnCore *business.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID
			id := web.Param(r, "agenda_id")

			if id != "" {
				agdID, err := uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}

				agd, err := agdCore.QueryGeneralAgendaByID(ctx, agdID)
				if err != nil {
					if errors.Is(err, agenda.ErrNotFound) {
						return response.NewError(err, http.StatusNotFound)
					}

					return fmt.Errorf("querybyid: agdID[%s]: %w", agdID, err)
				}
				bsn, err := bsnCore.QueryByID(ctx, agd.BusinessID)
				if err != nil {
					if errors.Is(err, business.ErrNotFound) {
						return response.NewError(err, http.StatusNotFound)
					}

					return fmt.Errorf("querybyid: bsnID[%s]: %w", agd.BusinessID, err)
				}

				userID = bsn.OwnerID
				setGeneralAgenda(ctx, agd)
				setBusiness(ctx, bsn)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func AuthorizeDailyAgenda(log *logger.Logger, ath *auth.Auth, agdCore *agenda.Core, bsnCore *business.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID
			id := web.Param(r, "agenda_id")

			if id != "" {
				agdID, err := uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}

				agd, err := agdCore.QueryDailyAgendaByID(ctx, agdID)
				if err != nil {
					if errors.Is(err, agenda.ErrNotFound) {
						return response.NewError(err, http.StatusNotFound)
					}

					return fmt.Errorf("querybyid: agdID[%s]: %w", agdID, err)
				}
				bsn, err := bsnCore.QueryByID(ctx, agd.BusinessID)
				if err != nil {
					if errors.Is(err, business.ErrNotFound) {
						return response.NewError(fmt.Errorf("you are not a business owner: %w", err), http.StatusNotFound)
					}

					return fmt.Errorf("querybyid: bsnID[%s]: %w", agd.BusinessID, err)
				}

				userID = bsn.OwnerID
				setDailyAgenda(ctx, agd)
				setBusiness(ctx, bsn)
			}

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			claims := auth.GetClaims(ctx)
			if err := ath.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
