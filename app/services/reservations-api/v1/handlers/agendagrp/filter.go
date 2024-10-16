package agendagrp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/google/uuid"
)

func parseGeneralAgendaFilter(r *http.Request) (agenda.GAQueryFilter, error) {
	const (
		filterByID         = "id"
		filterByBusinessID = "business_id"
	)

	values := r.URL.Query()
	var filter agenda.GAQueryFilter

	if agdID := values.Get(filterByID); agdID != "" {
		id, err := uuid.Parse(agdID)
		if err != nil {
			return agenda.GAQueryFilter{}, validate.NewFieldsError(filterByID, err)
		}

		filter.WithGenealAgendaID(id)
	}

	if bsnID := values.Get(filterByBusinessID); bsnID != "" {
		id, err := uuid.Parse(bsnID)
		if err != nil {
			return agenda.GAQueryFilter{}, validate.NewFieldsError(filterByBusinessID, err)
		}

		filter.WithBusinessID(id)
	}

	if err := filter.Validate(); err != nil {
		return agenda.GAQueryFilter{}, err
	}

	return filter, nil
}

func parseDailyAgendaFilter(r *http.Request) (agenda.DAQueryFilter, error) {
	const (
		filterByID         = "id"
		filterByBusinessID = "business_id"
		filterByDate       = "date"
		filterByFrom       = "from"
		filterByTo         = "to"
		filterByDays       = "days"
	)

	values := r.URL.Query()
	var filter agenda.DAQueryFilter

	if agdID := values.Get(filterByID); agdID != "" {
		id, err := uuid.Parse(agdID)
		if err != nil {
			return agenda.DAQueryFilter{}, validate.NewFieldsError(filterByID, err)
		}

		filter.WithDailyAgendaID(id)
	}

	if bsnID := values.Get(filterByBusinessID); bsnID != "" {
		id, err := uuid.Parse(bsnID)
		if err != nil {
			return agenda.DAQueryFilter{}, validate.NewFieldsError(filterByBusinessID, err)
		}

		filter.WithBusinessID(id)
	}

	if date := values.Get(filterByDate); date != "" {
		d, err := time.Parse(time.DateOnly, date)
		if err != nil {
			return agenda.DAQueryFilter{}, validate.NewFieldsError(filterByDate, err)
		}

		filter.WithDate(d.Format(time.DateOnly))
	}

	if from := values.Get(filterByFrom); from != "" {
		f, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return agenda.DAQueryFilter{}, validate.NewFieldsError(filterByFrom, err)
		}

		filter.WithFrom(f.Format(time.DateOnly))
	}

	if to := values.Get(filterByTo); to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err != nil {
			return agenda.DAQueryFilter{}, validate.NewFieldsError(filterByTo, err)
		}

		filter.WithTo(t.Format(time.DateOnly))
	}

	if days := values.Get(filterByDays); days != "" {
		d, err := strconv.Atoi(days)
		if err != nil {
			return agenda.DAQueryFilter{}, validate.NewFieldsError(filterByDays, err)
		}

		filter.WithDays(d)
	}

	if err := filter.Validate(); err != nil {
		return agenda.DAQueryFilter{}, err
	}

	return filter, nil
}
