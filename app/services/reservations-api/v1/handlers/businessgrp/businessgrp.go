package businessgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/business/web/v1/mid"
	"github.com/ameghdadian/service/business/web/v1/response"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper format")
)

type handlers struct {
	bsnCore *business.Core
	usrCore *user.Core
}

func newApp(bsnCore *business.Core, usrCore *user.Core) *handlers {
	return &handlers{
		bsnCore: bsnCore,
		usrCore: usrCore,
	}
}

func (h *handlers) executeUnderTransaction(ctx context.Context) (*handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		usrCore, err := h.usrCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		bsnCore, err := h.bsnCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &handlers{
			bsnCore: bsnCore,
			usrCore: usrCore,
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

	var app AppNewBusiness
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nb, err := toCoreNewBusiness(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	b, err := h.bsnCore.Create(ctx, nb)
	if err != nil {
		return errs.Newf(errs.Internal, "create: app[%+v]: %s", app, err)
	}

	return toAppBusiness(b)
}

func (h *handlers) update(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppUpdateBusiness
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	b, err := mid.GetBusiness(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "business missing in context: %s", err)
	}

	b, err = h.bsnCore.Update(ctx, b, toCoreUpdateBusiness(app))
	if err != nil {
		return errs.Newf(errs.Internal, "update: businessID[%s]: app[%+v]: %s", b.ID, app, err)
	}

	return toAppBusiness(b)
}

func (h *handlers) delete(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	b, err := mid.GetBusiness(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "business missing in context: %s", err)
	}

	if err := h.bsnCore.Delete(ctx, b); err != nil {
		return errs.Newf(errs.Internal, "delete: businessID[%s]: %s", b.ID, err)
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
		return errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err.(*errs.Error)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, business.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	bsns, err := h.bsnCore.Query(ctx, filter, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := h.bsnCore.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return response.NewPageDocument(toAppBusinesses(bsns), total, page)

}

func (h *handlers) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	bsnID, err := uuid.Parse(web.Param(r, "business_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	b, err := h.bsnCore.QueryByID(ctx, bsnID)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return errs.New(errs.NotFound, err)
		default:
			return errs.Newf(errs.Internal, "querybyid: businessID[%s]: %s", bsnID, err)
		}
	}

	return toAppBusiness(b)
}
