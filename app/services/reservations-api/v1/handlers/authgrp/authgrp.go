package authgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/ameghdadian/service/foundation/web"
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

func (h *Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	kid := web.Param(r, "kid")
	if kid == "" {
		return validate.NewFieldsError("kid", errors.New("missing kid"))
	}

	claims := auth.GetClaims(ctx)

	token, err := h.auth.GenerateToken(kid, claims)
	if err != nil {
		return fmt.Errorf("generatetoken: %w", err)
	}

	return web.Respond(ctx, w, toToken(token), http.StatusOK)
}

func (h *Handlers) Authenticate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// The middleware handles the authentication. So when code gets to this
	// handler, authentication is passed.

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return auth.NewAuthError("%s", err)
	}

	resp := authenticateResp{
		UserID: userID,
		Claims: auth.GetClaims(ctx),
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}
