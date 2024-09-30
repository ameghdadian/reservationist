package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/response"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/golang-jwt/jwt/v4"
)

type Handlers struct {
	user *user.Core
	auth *auth.Auth
}

func New(user *user.Core, auth *auth.Auth) *Handlers {
	return &Handlers{
		user: user,
		auth: auth,
	}
}

func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		user, err := h.user.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &Handlers{
			user: user,
			auth: h.auth,
		}

		return h, nil
	}

	return h, nil
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppNewUser
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	nc, err := toCoreNewUser(app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	usr, err := h.user.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmailOrPhoneNo) {
			return response.NewError(err, http.StatusConflict)
		}
		return fmt.Errorf("create: usr[%+v]: %w", usr, err)
	}

	return web.Respond(ctx, w, toAppUser(usr), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppUpdateUser
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	userID := auth.GetUserID(ctx)

	usr, err := h.user.QueryByID(ctx, userID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: userID[%s]: %w", userID, err)
		}
	}

	uu, err := toCoreUpdateUser(app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	usr, err = h.user.Update(ctx, usr, uu)
	if err != nil {
		return fmt.Errorf("update: userID[%s] uu[%+v]: %w", userID, uu, err)
	}

	return web.Respond(ctx, w, toAppUser(usr), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	userID := auth.GetUserID(ctx)

	usr, err := h.user.QueryByID(ctx, userID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: userID[%s]: %w", userID, err)
		}
	}

	if err := h.user.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: userID[%v]: %w", userID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := page.Parse(r)
	if err != nil {
		return err
	}

	filter, err := parseFilter(r)
	if err != nil {
		return err
	}

	orderBy, err := parseOrder(r)
	if err != nil {
		return err
	}

	users, err := h.user.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.user.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, response.NewPageDocument(toAppUsers(users), total, page.Number, page.RowsPerPage), http.StatusOK)
}

func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := auth.GetUserID(ctx)

	usr, err := h.user.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: id[%s]: %w", id, err)
		}
	}

	return web.Respond(ctx, w, toAppUser(usr), http.StatusOK)
}

func (h *Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	kid := web.Param(r, "kid")
	if kid == "" {
		return validate.NewFieldsError("kid", errors.New("missing kid"))
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		return auth.NewAuthError("must provide email and password in Basic auth")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return auth.NewAuthError("invalid email format")
	}

	usr, err := h.user.Authenticate(ctx, *addr, pass)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		case errors.Is(err, user.ErrAuthenticationFailure):
			return auth.NewAuthError(err.Error())
		default:
			return fmt.Errorf("authenticate: %w", err)
		}
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

	token, err := h.auth.GenerateToken(kid, claims)
	if err != nil {
		return fmt.Errorf("generatetoken: %w", err)
	}

	return web.Respond(ctx, w, toToken(token), http.StatusOK)
}
