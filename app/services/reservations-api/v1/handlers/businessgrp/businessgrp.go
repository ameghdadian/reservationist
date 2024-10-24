package businessgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/business/web/v1/mid"
	"github.com/ameghdadian/service/business/web/v1/response"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper format")
)

type Handlers struct {
	bsnCore *business.Core
	usrCore *user.Core
}

func New(bsnCore *business.Core, usrCore *user.Core) *Handlers {
	return &Handlers{
		bsnCore: bsnCore,
		usrCore: usrCore,
	}
}

func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		usrCore, err := h.usrCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		bsnCore, err := h.bsnCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &Handlers{
			bsnCore: bsnCore,
			usrCore: usrCore,
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

	var app AppNewBusiness
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	nb, err := toCoreNewBusiness(app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	b, err := h.bsnCore.Create(ctx, nb)
	if err != nil {
		return fmt.Errorf("create: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppBusiness(b), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppUpdateBusiness
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	bsnId, err := uuid.Parse(web.Param(r, "business_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	b, err := mid.GetBusiness(ctx)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: businessID[%s]: %w", bsnId, err)
		}
	}

	b, err = h.bsnCore.Update(ctx, b, toCoreUpdateBusiness(app))
	if err != nil {
		return fmt.Errorf("update: businessID[%s]: app[%+v]: %w", bsnId, app, err)
	}

	return web.Respond(ctx, w, toAppBusiness(b), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	bsnID, err := uuid.Parse(web.Param(r, "business_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	b, err := mid.GetBusiness(ctx)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return response.NewError(err, http.StatusNoContent)
		default:
			return fmt.Errorf("querybyid: businessID[%s]: %w", bsnID, err)
		}
	}

	if err := h.bsnCore.Delete(ctx, b); err != nil {
		return fmt.Errorf("delete: businessID[%s]: %w", bsnID, err)
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

	bsns, err := h.bsnCore.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.bsnCore.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, response.NewPageDocument(toAppBusinesses(bsns), total, page.Number, page.RowsPerPage), http.StatusOK)

}

func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	bsnID, err := uuid.Parse(web.Param(r, "business_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	b, err := h.bsnCore.QueryByID(ctx, bsnID)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: businessID[%s]: %w", bsnID, err)
		}
	}

	return web.Respond(ctx, w, toAppBusiness(b), http.StatusOK)
}
