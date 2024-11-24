package mid

import (
	"context"
	"net/http"
	"net/mail"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
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

func Bearer(ath *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := ath.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewAuthError("%s", err)
			}

			if claims.Subject == "" {
				return auth.NewAuthError("authorize: you are not authorized for that action, no claims")
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return auth.NewAuthError("parsing subject: %s", err)
			}

			ctx = auth.SetUserID(ctx, subjectID)
			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Basic processes basic authentication logic.
func Basic(ath *auth.Auth, usrCore *user.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			email, pass, ok := r.BasicAuth()
			if !ok {
				return auth.NewAuthError("must provide email and password in Basic auth")
			}

			addr, err := mail.ParseAddress(email)
			if err != nil {
				return auth.NewAuthError("invalid email format")
			}

			usr, err := usrCore.Authenticate(ctx, *addr, pass)
			if err != nil {
				return auth.NewAuthError(err.Error())
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
				return auth.NewAuthError("parsing subject: %s", err)
			}

			ctx = auth.SetUserID(ctx, subjectID)
			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
