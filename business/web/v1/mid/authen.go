package mid

import (
	"context"
	"net/http"
	"net/mail"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func Authenticate(a *auth.Auth) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, r *http.Request) web.Encoder {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			ctx = auth.SetClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

func Bearer(ath *auth.Auth) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			claims, err := ath.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			if claims.Subject == "" {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, no claims")
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
			}

			ctx = auth.SetUserID(ctx, subjectID)
			ctx = auth.SetClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

// Basic processes basic authentication logic.
func Basic(ath *auth.Auth, usrCore *user.Core) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			email, pass, ok := r.BasicAuth()
			if !ok {
				return errs.Newf(errs.Unauthenticated, "invalid Basic auth")
			}

			addr, err := mail.ParseAddress(email)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			usr, err := usrCore.Authenticate(ctx, *addr, pass)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   usr.ID.String(),
					Issuer:    "service project",
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				},
				Roles: usr.Roles,
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
			}

			ctx = auth.SetUserID(ctx, subjectID)
			ctx = auth.SetClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}
