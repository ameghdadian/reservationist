package appointmentgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/appointment"
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
	aptCore *appointment.Core
	agdCore *agenda.Core
}

func newApp(aptCore *appointment.Core, agdCore *agenda.Core) *handlers {
	return &handlers{
		aptCore: aptCore,
		agdCore: agdCore,
	}
}

func (h *handlers) executeUnderTransaction(ctx context.Context) (*handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		aptCore, err := h.aptCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}
		agdCore, err := h.agdCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &handlers{
			aptCore: aptCore,
			agdCore: agdCore,
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

	var app AppNewAppointment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	na, err := toCoreNewAppointment(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Check whether appointment conforms with daily agenda.
	err = h.agdCore.ConformDailyAgendaBoundary(ctx, na.BusinessID, na.ScheduledOn)
	if err != nil {
		if errors.Is(err, agenda.ErrNoDailyAgenda) {
			// If doesn't conform with daily agenda, check with the general agenda to see any match.
			if err = h.agdCore.ConformGeneralAgendaBoundary(ctx, na.BusinessID, na.ScheduledOn); err != nil {
				return errs.New(errs.InvalidArgument, err)
			}
		} else {
			return errs.New(errs.InvalidArgument, err)
		}
	}

	apt, err := h.aptCore.Create(ctx, na)
	if err != nil {
		return errs.Newf(errs.Internal, "create: app[%+v]: %s", app, err)
	}

	return toAppAppointment(apt)
}

func (h *handlers) update(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppUpdateAppointment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	aptID, err := uuid.Parse(web.Param(r, "appointment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	apt, err := mid.GetAppointment(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "appointment missing in context: %s", err)
	}

	uapt, err := toCoreUpdateAppointment(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	apt, err = h.aptCore.Update(ctx, apt, uapt)
	if err != nil {
		return errs.Newf(errs.Internal, "update: appointmentID[%s] uapt[%+v]: %s", aptID, uapt, err)
	}

	return toAppAppointment(apt)
}

func (h *handlers) delete(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	aptID, err := uuid.Parse(web.Param(r, "appointment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	apt, err := mid.GetAppointment(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "appointment missing in context: %s", err)
	}

	if err := h.aptCore.Delete(ctx, apt); err != nil {
		return errs.Newf(errs.Internal, "delete: appointmentID[%s]: %s", aptID, err)
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
		return errs.New(errs.InvalidArgument, err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err.(*errs.Error)
	}

	// TODO: What's the returned status when a Field validation error occurs?
	orderBy, err := order.Parse(orderByFields, qp.OrderBy, appointment.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	apts, err := h.aptCore.Query(ctx, filter, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := h.aptCore.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return response.NewPageDocument(toAppAppointments(apts), total, page)
}

func (h *handlers) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	_, err := uuid.Parse(web.Param(r, "appointment_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	apt, err := mid.GetAppointment(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "appointment missing in context: %s", err)
	}

	return toAppAppointment(apt)
}
