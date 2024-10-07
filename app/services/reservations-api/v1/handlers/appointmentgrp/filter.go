package appointmentgrp

import (
	"net/http"
	"time"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(r *http.Request) (appointment.QueryFilter, error) {
	const (
		filterByID               = "appointment_id"
		filterByBusinessID       = "business_id"
		filterByUserID           = "user_id"
		filterByStatus           = "status"
		filterByScheduledOn      = "scheduled_on"
		filterByStartCreatedDate = "start_created_date"
		filterByEndCreatedDate   = "end_created_date"
	)

	values := r.URL.Query()

	var filter appointment.QueryFilter

	if aptID := values.Get(filterByID); aptID != "" {
		id, err := uuid.Parse(aptID)
		if err != nil {
			return appointment.QueryFilter{}, validate.NewFieldsError(filterByID, err)
		}
		filter.WithAppointmentID(id)
	}

	if bsnID := values.Get(filterByBusinessID); bsnID != "" {
		id, err := uuid.Parse(bsnID)
		if err != nil {
			return appointment.QueryFilter{}, validate.NewFieldsError(filterByBusinessID, err)
		}
		filter.WithBusinessID(id)
	}

	if usrID := values.Get(filterByUserID); usrID != "" {
		id, err := uuid.Parse(usrID)
		if err != nil {
			return appointment.QueryFilter{}, validate.NewFieldsError(filterByUserID, err)
		}
		filter.WithUserID(id)
	}

	if status := values.Get(filterByStatus); status != "" {
		st, err := appointment.ParseStatus(status)
		if err != nil {
			return appointment.QueryFilter{}, validate.NewFieldsError(filterByStatus, err)
		}
		filter.WithStatus(st)
	}

	if sch := values.Get(filterByScheduledOn); sch != "" {
		t, err := time.Parse(time.RFC3339, sch)
		if err != nil {
			return appointment.QueryFilter{}, validate.NewFieldsError(filterByScheduledOn, err)
		}
		filter.WithScheduledOn(t)
	}

	if startDate := values.Get(filterByStartCreatedDate); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			return appointment.QueryFilter{}, validate.NewFieldsError(filterByStartCreatedDate, err)
		}
		filter.WithStartCreatedDate(t)
	}

	if endDate := values.Get(filterByEndCreatedDate); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			return appointment.QueryFilter{}, validate.NewFieldsError(filterByEndCreatedDate, err)
		}
		filter.WithStartCreatedDate(t)
	}

	if err := filter.Validate(); err != nil {
		return appointment.QueryFilter{}, err
	}

	return filter, nil
}
