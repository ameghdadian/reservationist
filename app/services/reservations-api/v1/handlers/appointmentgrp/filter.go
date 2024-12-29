package appointmentgrp

import (
	"net/http"
	"time"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/google/uuid"
)

func parseQueryParams(r *http.Request) (queryParams, error) {
	values := r.URL.Query()

	filter := queryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("rows"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("appointment_id"),
		BusinessID:       values.Get("business_id"),
		UserID:           values.Get("user_id"),
		Status:           values.Get("status"),
		ScheduledOn:      values.Get("scheduled_on"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
	}

	return filter, nil
}

func parseFilter(qp queryParams) (appointment.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter appointment.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.WithAppointmentID(id)
		default:
			fieldErrors.Add("appointment_id", err)
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

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		switch err {
		case nil:

			filter.WithUserID(id)
		default:
			fieldErrors.Add("user_id", err)
		}
	}

	if qp.Status != "" {
		st, err := appointment.ParseStatus(qp.Status)
		switch err {
		case nil:
			filter.WithStatus(st)
		default:
			fieldErrors.Add("status", err)
		}
	}

	if qp.ScheduledOn != "" {
		t, err := time.Parse(time.RFC3339, qp.ScheduledOn)
		switch err {
		case nil:
			filter.WithScheduledOn(t)
		default:
			fieldErrors.Add("scheduled_on", err)
		}
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		switch err {
		case nil:
			filter.WithStartCreatedDate(t)
		default:
			fieldErrors.Add("start_created_date", err)
		}
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		switch err {
		case nil:
			filter.WithStartCreatedDate(t)
		default:
			fieldErrors.Add("end_created_date", err)
		}
	}

	if err := filter.Validate(); err != nil {
		fieldErrors.Add("filter validation", err)
	}

	if fieldErrors != nil {
		return appointment.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
