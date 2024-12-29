package agendagrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/business/web/v1/auth"
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
	agdCore *agenda.Core
	bsnCore *business.Core
}

func newApp(agdCore *agenda.Core, bsnCore *business.Core) *handlers {
	return &handlers{
		agdCore: agdCore,
		bsnCore: bsnCore,
	}
}

func (h *handlers) executeUnderTransaction(ctx context.Context) (*handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		agdCore, err := h.agdCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		bsnCore, err := h.bsnCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		return &handlers{
			agdCore: agdCore,
			bsnCore: bsnCore,
		}, nil
	}

	return h, nil
}

func (h *handlers) createGeneralAgenda(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppNewGeneralAgenda
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nAgd, err := toCoreNewGeneralAgenda(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	bsn, err := h.bsnCore.QueryByID(ctx, nAgd.BusinessID)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return errs.New(errs.NotFound, err)
		default:
			return errs.Newf(errs.Internal, "create general agenda: app[%+v]: %s", app, err)
		}
	}

	usrClaimID := auth.GetClaims(ctx).Subject
	if usrClaimID != bsn.OwnerID.String() {
		return errs.Newf(errs.PermissionDenied, "you don't have the persmission for this action: %s", auth.ErrForbidden)
	}

	gAgd, err := h.agdCore.CreateGeneralAgenda(ctx, nAgd)
	if err != nil {
		return errs.Newf(errs.Internal, "create general agenda: app[%+v]: %s", app, err)
	}

	return toAppGeneralAgenda(gAgd)
}

func (h *handlers) updateGeneralAgenda(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppUpdateGeneralAgenda
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	gAgdID, err := uuid.Parse(web.Param(r, "agenda_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	agd, err := mid.GetGeneralAgenda(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "general agenda missing in context: %s", err)
	}

	uAgd, err := toCoreUpdateGeneralAgenda(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	agd, err = h.agdCore.UpdateGenralAgenda(ctx, agd, uAgd)
	if err != nil {
		return errs.Newf(errs.Internal, "update: generalAgendaID[%s]: %s", gAgdID, err)
	}

	return toAppGeneralAgenda(agd)
}

func (h *handlers) deleteGeneralAgenda(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	agdID, err := uuid.Parse(web.Param(r, "agenda_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	agd, err := mid.GetGeneralAgenda(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "general agenda missing in context: %s", err)
	}

	if err := h.agdCore.DeleteGeneralAgenda(ctx, agd); err != nil {
		return errs.Newf(errs.Internal, "delete: generalAgendaID[%s]: %s", agdID, err)
	}

	return nil
}

func (h *handlers) queryGeneralAgenda(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseGeneralAgendaQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	filter, err := parseGeneralAgendaFilter(qp)
	if err != nil {
		return err.(*errs.Error)
	}

	orderBy, err := order.Parse(generalAgendaOrderByFields, qp.OrderBy, agenda.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	agds, err := h.agdCore.QueryGeneralAgenda(ctx, filter, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := h.agdCore.CountGeneralAgenda(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	return response.NewPageDocument(toAppGeneralAgendaSlice(agds), total, page)
}

func (h *handlers) queryGeneralAgendaByID(ctx context.Context, r *http.Request) web.Encoder {
	agdID, err := uuid.Parse(web.Param(r, "agenda_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	agd, err := h.agdCore.QueryGeneralAgendaByID(ctx, agdID)
	if err != nil {
		return errs.Newf(errs.Internal, "querygeneralagendabyagendaid: agdID[%s]: %s", agdID, err)
	}

	return toAppGeneralAgenda(agd)
}

// ---------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------

func (h *handlers) createDailyAgenda(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppNewDailyAgenda
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	nAgd, err := toCoreNewDailyAgenda(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	bsn, err := h.bsnCore.QueryByID(ctx, nAgd.BusinessID)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return errs.New(errs.NotFound, err)
		default:
			return errs.Newf(errs.Internal, "create general agenda: app[%+v]: %s", app, err)
		}
	}

	usrClaimID := auth.GetClaims(ctx).Subject
	if usrClaimID != bsn.OwnerID.String() {
		return errs.Newf(errs.PermissionDenied, "you don't have the persmission for this action: %s", auth.ErrForbidden)
	}

	gAgd, err := h.agdCore.CreateDailyAgenda(ctx, nAgd)
	if err != nil {
		return errs.Newf(errs.Internal, "create daily agenda: app[%+v]: %s", app, err)
	}

	return toAppDailyAgenda(gAgd)
}

func (h *handlers) updateDailyAgenda(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	var app AppUpdateDailyAgenda
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	gAgdID, err := uuid.Parse(web.Param(r, "agenda_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	agd, err := mid.GetDailyAgenda(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "daily agenda missing in context: %s", err)
	}

	uAgd, err := toCoreUpdateDailyAgenda(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	agd, err = h.agdCore.UpdateDailyAgenda(ctx, agd, uAgd)
	if err != nil {
		return errs.Newf(errs.Internal, "update: dailyAgendaID[%s]: %s", gAgdID, err)
	}

	return toAppDailyAgenda(agd)
}

func (h *handlers) deleteDailyAgenda(ctx context.Context, r *http.Request) web.Encoder {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	agdID, err := uuid.Parse(web.Param(r, "agenda_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	agd, err := mid.GetDailyAgenda(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "daily agenda missing in context: %s", err)
	}

	if err := h.agdCore.DeleteDailyAgenda(ctx, agd); err != nil {
		return errs.Newf(errs.Internal, "delete: dailyAgendaID[%s]: %s", agdID, err)
	}

	return nil
}

func (h *handlers) queryDailyAgenda(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseDailyAgendaQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	filter, err := parseDailyAgendaFilter(qp)
	if err != nil {
		return err.(*errs.Error)
	}

	orderBy, err := order.Parse(dailyAgendaOrderByFields, qp.OrderBy, agenda.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	agds, err := h.agdCore.QueryDailyAgenda(ctx, filter, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := h.agdCore.CountDailyAgenda(ctx, agenda.DAQueryFilter{})
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return response.NewPageDocument(toAppDailyAgendaSlice(agds), total, page)
}

func (h *handlers) queryDailyAgendaByID(ctx context.Context, r *http.Request) web.Encoder {
	agdID, err := uuid.Parse(web.Param(r, "agenda_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, ErrInvalidID)
	}

	agd, err := h.agdCore.QueryDailyAgendaByID(ctx, agdID)
	if err != nil {
		return errs.Newf(errs.Internal, "querydailyagendabyid: agdID[%s]: %s", agdID, err)
	}

	return toAppDailyAgenda(agd)

}
