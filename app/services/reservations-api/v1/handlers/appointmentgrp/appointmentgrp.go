package appointmentgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/business/web/v1/response"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper format")
)

type Handlers struct {
	aptCore *appointment.Core
}

func New(aptCore *appointment.Core) *Handlers {
	return &Handlers{
		aptCore: aptCore,
	}
}

func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		aptCore, err := h.aptCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &Handlers{
			aptCore: aptCore,
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

	var app AppNewAppointment
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	na, err := toCoreNewAppointment(app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	apt, err := h.aptCore.Create(ctx, na)
	if err != nil {
		return fmt.Errorf("create: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppAppointment(apt), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppUpdateAppointment
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	aptID, err := uuid.Parse(web.Param(r, "appointment_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	apt, err := h.aptCore.QueryByID(ctx, aptID)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: appointmentid[%s]: %w", aptID, err)
		}
	}

	uapt, err := toCoreUpdateAppointment(app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	apt, err = h.aptCore.Update(ctx, apt, uapt)
	if err != nil {
		return fmt.Errorf("update: appointmentID[%s] uapt[%+v]: %w", aptID, uapt, err)
	}

	return web.Respond(ctx, w, toAppAppointment(apt), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	aptID, err := uuid.Parse(web.Param(r, "appointment_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	apt, err := h.aptCore.QueryByID(ctx, aptID)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: appointmentid[%s]: %w", aptID, err)
		}
	}

	if err := h.aptCore.Delete(ctx, apt); err != nil {
		return fmt.Errorf("delete: appointmentID[%s]: %w", aptID, err)
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

	// TODO: What's the returned status when a Field validation error occurs?
	orderBy, err := parseOrder(r)
	if err != nil {
		return err
	}

	apts, err := h.aptCore.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.aptCore.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, response.NewPageDocument(toAppAppointments(apts), total, page.Number, page.RowsPerPage), http.StatusOK)
}

func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	aptID, err := uuid.Parse(web.Param(r, "appointment_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	apt, err := h.aptCore.QueryByID(ctx, aptID)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: appointmentid[%s]: %w", aptID, err)
		}
	}

	return web.Respond(ctx, w, toAppAppointment(apt), http.StatusOK)
}
