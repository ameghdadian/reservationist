package authgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/web"
)

type handlers struct {
	user *user.Core
	auth *auth.Auth
}

func newApp(user *user.Core, auth *auth.Auth) *handlers {
	return &handlers{
		user: user,
		auth: auth,
	}
}

func (h *handlers) token(ctx context.Context, r *http.Request) web.Encoder {
	kid := web.Param(r, "kid")
	if kid == "" {
		return errs.NewFieldErrors("kid", errors.New("missing kid"))
	}

	claims := auth.GetClaims(ctx)

	token, err := h.auth.GenerateToken(kid, claims)
	if err != nil {
		return errs.Newf(errs.Internal, "generatetoken: %s", err)
	}

	return toToken(token)
}

func (h *handlers) authenticate(ctx context.Context, r *http.Request) web.Encoder {
	// The middleware handles the authentication. So when code gets to this
	// handler, authentication is passed.

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "%s", err)
	}

	resp := authenticateResp{
		UserID: userID,
		Claims: auth.GetClaims(ctx),
	}

	return resp
}
