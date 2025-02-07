package agendagrp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/google/uuid"
)

func parseGeneralAgendaQueryParams(r *http.Request) (generalAgendaQueryParams, error) {
	values := r.URL.Query()

	filter := generalAgendaQueryParams{
		Page:       values.Get("page"),
		Rows:       values.Get("rows"),
		OrderBy:    values.Get("orderBy"),
		ID:         values.Get("id"),
		BusinessID: values.Get("business_id"),
	}

	return filter, nil
}

func parseDailyAgendaQueryParams(r *http.Request) (dailyAgendaQueryParams, error) {
	values := r.URL.Query()

	filter := dailyAgendaQueryParams{
		Page:       values.Get("page"),
		Rows:       values.Get("rows"),
		OrderBy:    values.Get("orderBy"),
		ID:         values.Get("id"),
		BusinessID: values.Get("business_id"),
		Date:       values.Get("date"),
		From:       values.Get("from"),
		To:         values.Get("to"),
		Days:       values.Get("days"),
	}

	return filter, nil
}

func parseGeneralAgendaFilter(qp generalAgendaQueryParams) (agenda.GAQueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter agenda.GAQueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.WithGenealAgendaID(id)
		default:
			fieldErrors.Add("id", err)
		}
	}

	if qp.BusinessID != "" {
		id, err := uuid.Parse(qp.BusinessID)
		switch err {
		case nil:
			filter.WithBusinessID(id)
		default:
			fieldErrors.Add("business_id", err)
		}
	}

	if err := filter.Validate(); err != nil {
		fieldErrors.Add("filter validation", err)
	}

	if fieldErrors != nil {
		return agenda.GAQueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}

func parseDailyAgendaFilter(qp dailyAgendaQueryParams) (agenda.DAQueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter agenda.DAQueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.WithDailyAgendaID(id)
		default:
			fieldErrors.Add("id", err)
		}
	}

	if qp.BusinessID != "" {
		id, err := uuid.Parse(qp.BusinessID)
		switch err {
		case nil:
			filter.WithBusinessID(id)
		default:
			fieldErrors.Add("business_id", err)
		}
	}

	if qp.Date != "" {
		d, err := time.Parse(time.DateOnly, qp.Date)
		switch err {
		case nil:
			filter.WithDate(d.UTC())
		default:
			fieldErrors.Add("date", err)
		}
	}

	if qp.From != "" {
		f, err := time.Parse(time.DateOnly, qp.From)
		switch err {
		case nil:
			filter.WithFrom(f.Format(time.DateOnly))
		default:
			fieldErrors.Add("from", err)
		}
	}

	if qp.To != "" {
		t, err := time.Parse(time.DateOnly, qp.To)
		switch err {
		case nil:
			filter.WithTo(t.Format(time.DateOnly))
		default:
			fieldErrors.Add("to", err)
		}
	}

	if qp.Days != "" {
		d, err := strconv.Atoi(qp.Days)
		switch err {
		case nil:
			filter.WithDays(d)
		default:
			fieldErrors.Add("days", err)
		}
	}

	if err := filter.Validate(); err != nil {
		return agenda.DAQueryFilter{}, errs.NewFieldErrors("filter validation", err)
	}

	if fieldErrors != nil {
		return agenda.DAQueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
