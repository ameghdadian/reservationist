package usergrp

import (
	"net/http"
	"net/mail"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/google/uuid"
)

func parseQueryParams(r *http.Request) (queryParams, error) {
	values := r.URL.Query()

	filter := queryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("rows"),
		OrderBy:          values.Get("orderBy"),
		UserID:           values.Get("user_id"),
		Email:            values.Get("email"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
		Name:             values.Get("name"),
		PhoneNumber:      values.Get("phone_number"),
	}

	return filter, nil
}

func parseFilter(qp queryParams) (user.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter user.QueryFilter

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		switch err {
		case nil:
			filter.WithUserID(id)
		default:
			fieldErrors.Add("user_id", err)
		}
	}

	if qp.Email != "" {
		addr, err := mail.ParseAddress(qp.Email)
		switch err {
		case nil:
			filter.WithEmail(*addr)
		default:
			fieldErrors.Add("email", err)
		}
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		switch err {
		case nil:
			filter.WithStartDateCreated(t)
		default:
			fieldErrors.Add("start_created_date", err)
		}
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		switch err {
		case nil:
			filter.WithEndDateCreated(t)
		default:
			fieldErrors.Add("end_created_date", err)
		}
	}

	if qp.Name != "" {
		filter.WithName(qp.Name)
	}

	if qp.PhoneNumber != "" {
		num, err := user.ParsePhoneNumber(qp.PhoneNumber)
		switch err {
		case nil:
			filter.WithPhoneNumber(num)
		default:
			fieldErrors.Add("phone_number", err)
		}
	}

	if err := filter.Validate(); err != nil {
		fieldErrors.Add("filter validation", err)
	}

	if fieldErrors != nil {
		return user.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
