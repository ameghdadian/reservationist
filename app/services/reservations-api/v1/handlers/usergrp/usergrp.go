package usergrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/response"
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

func (h *handlers) executeUnderTransaction(ctx context.Context) (*handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		user, err := h.user.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &handlers{
			user: user,
			auth: h.auth,
		}

		return h, nil
	}

	return h, nil
}

func (h *handlers) create(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppNewUser
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nc, err := toCoreNewUser(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err := h.user.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmailOrPhoneNo) {
			return errs.New(errs.Aborted, err)
		}
		return errs.Newf(errs.Internal, "create: usr[%+v]: %s", app, err)
	}

	return toAppUser(usr)
}

func (h *handlers) update(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppUpdateUser
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "update: %s", err)
	}

	usr, err := h.user.QueryByID(ctx, userID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return errs.New(errs.NotFound, err)
		default:
			return errs.Newf(errs.Internal, "querybyid: userID[%s]: %s", userID, err)
		}
	}

	uu, err := toCoreUpdateUser(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err = h.user.Update(ctx, usr, uu)
	if err != nil {
		return errs.Newf(errs.Internal, "update: userID[%s] uu[%+v]: %s", userID, uu, err)
	}

	return toAppUser(usr)
}

func (h *handlers) delete(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	usr, err := h.user.QueryByID(ctx, userID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return errs.New(errs.NotFound, err)
		default:
			return errs.Newf(errs.Internal, "querybyid: userID[%s]: %s", userID, err)
		}
	}

	if err := h.user.Delete(ctx, usr); err != nil {
		return errs.Newf(errs.Internal, "delete: userID[%v]: %s", userID, err)
	}

	return nil
}

func (h *handlers) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		// TODO: Does FieldErrors end up being 500 error? Are we handling it in web.Respond
		// anywhere else?
		return errs.New(errs.InvalidArgument, errs.NewFieldErrors("page", err))
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err.(errs.FieldErrors))
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, user.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	users, err := h.user.Query(ctx, filter, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := h.user.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return response.NewPageDocument(toAppUsers(users), total, page)
}

func (h *handlers) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id, err := auth.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	usr, err := h.user.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return errs.New(errs.NotFound, err)
		default:
			return errs.Newf(errs.Internal, "querybyid: id[%s]: %s", id, err)
		}
	}

	return toAppUser(usr)
}
