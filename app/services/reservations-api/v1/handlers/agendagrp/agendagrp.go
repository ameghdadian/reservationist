package agendagrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/business"
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
	agdCore *agenda.Core
	bsnCore *business.Core
}

func New(agdCore *agenda.Core, bsnCore *business.Core) *Handlers {
	return &Handlers{
		agdCore: agdCore,
		bsnCore: bsnCore,
	}
}

func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		agdCore, err := h.agdCore.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		return &Handlers{
			agdCore: agdCore,
		}, nil
	}

	return h, nil
}

func (h *Handlers) CreateGeneralAgenda(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppNewGeneralAgenda
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	nAgd, err := toCoreNewGeneralAgenda(app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	gAgd, err := h.agdCore.CreateGeneralAgenda(ctx, nAgd)
	if err != nil {
		return fmt.Errorf("create general agenda: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppGeneralAgenda(gAgd), http.StatusCreated)
}

func (h *Handlers) UpdateGeneralAgenda(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppUpdateGeneralAgenda
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	gAgdID, err := uuid.Parse(web.Param(r, "general_agenda_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	agd, err := h.agdCore.QueryGeneralAgendaByID(ctx, gAgdID)
	if err != nil {
		switch {
		case errors.Is(err, agenda.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: generalAgendaID[%s]: %w", gAgdID, err)
		}
	}

	uAgd, err := toCoreUpdateGeneralAgenda(app)
	if err != nil {
		return err
	}

	agd, err = h.agdCore.UpdateGenralAgenda(ctx, agd, uAgd)
	if err != nil {
		return fmt.Errorf("update: generalAgendaID[%s]: %w", gAgdID, err)
	}

	return web.Respond(ctx, w, toAppGeneralAgenda(agd), http.StatusOK)
}

func (h *Handlers) DeleteGeneralAgenda(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	agdID, err := uuid.Parse(web.Param(r, "general_agenda_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	agd, err := h.agdCore.QueryGeneralAgendaByID(ctx, agdID)
	if err != nil {
		switch {
		case errors.Is(err, agenda.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: generalAgendaID[%s]: %w", agdID, err)
		}
	}

	if err := h.agdCore.DeleteGeneralAgenda(ctx, agd); err != nil {
		return fmt.Errorf("delete: generalAgendaID[%s]: %w", agdID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) QueryGeneralAgendaByBusinessID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	bsnID, err := uuid.Parse(web.Param(r, "business_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	_, err = h.bsnCore.QueryByID(ctx, bsnID)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return response.NewError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("querybyid: bsnID[%s]: %w", bsnID, err)
		}
	}

	agd, err := h.agdCore.QueryGeneralAgendaByBusinessID(ctx, bsnID)
	if err != nil {
		return fmt.Errorf("querygeneralagendabybusinessid: bsnID[%s]: %w", bsnID, err)
	}

	return web.Respond(ctx, w, toAppGeneralAgenda(agd), http.StatusOK)
}

func (h *Handlers) CreateDailyAgenda(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppNewDailyAgenda
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	nAgd, err := toCoreNewDailyAgenda(app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	gAgd, err := h.agdCore.CreateDailyAgenda(ctx, nAgd)
	if err != nil {
		return fmt.Errorf("create daily agenda: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppDailyAgenda(gAgd), http.StatusCreated)
}

func (h *Handlers) UpdateDailyAgenda(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppUpdateDailyAgenda
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	gAgdID, err := uuid.Parse(web.Param(r, "daily_agenda_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	agd, err := h.agdCore.QueryDailyAgendaByID(ctx, gAgdID)
	if err != nil {
		switch {
		case errors.Is(err, agenda.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: dailyAgendaID[%s]: %w", gAgdID, err)
		}
	}

	uAgd, err := toCoreUpdateDailyAgenda(app)
	if err != nil {
		return err
	}

	agd, err = h.agdCore.UpdateDailyAgenda(ctx, agd, uAgd)
	if err != nil {
		return fmt.Errorf("update: dailyAgendaID[%s]: %w", gAgdID, err)
	}

	return web.Respond(ctx, w, toAppDailyAgenda(agd), http.StatusOK)
}

func (h *Handlers) DeleteDailyAgenda(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	agdID, err := uuid.Parse(web.Param(r, "daily_agenda_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	agd, err := h.agdCore.QueryDailyAgendaByID(ctx, agdID)
	if err != nil {
		switch {
		case errors.Is(err, agenda.ErrNotFound):
			return response.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: dailyAgendaID[%s]: %w", agdID, err)
		}
	}

	if err := h.agdCore.DeleteDailyAgenda(ctx, agd); err != nil {
		return fmt.Errorf("delete: dailyAgendaID[%s]: %w", agdID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) QueryDailyAgenda(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := page.Parse(r)
	if err != nil {
		return err
	}

	filter, err := parseDailyAgendaFilter(r)
	if err != nil {
		return err
	}

	orderBy, err := parseDailyAgendaOrder(r)
	if err != nil {
		return err
	}

	// TODO: /businesses/<business_id>/daily-agenda
	bsnID, err := uuid.Parse(web.Param(r, "business_id"))
	if err != nil {
		return response.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	_, err = h.bsnCore.QueryByID(ctx, bsnID)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrNotFound):
			return response.NewError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("querybyid: bsnID[%s]: %w", bsnID, err)
		}
	}

	agds, err := h.agdCore.QueryDailyAgenda(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.agdCore.CountDailyAgenda(ctx, agenda.DAQueryFilter{})
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, response.NewPageDocument(toAppDailyAgendaSlice(agds), total, page.Number, page.RowsPerPage), http.StatusOK)
}
