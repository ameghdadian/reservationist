package businessgrp

import (
	"net/http"
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/google/uuid"
)

func parseQueryParams(r *http.Request) (queryParams, error) {
	values := r.URL.Query()

	filter := queryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("row"),
		OrderBy:          values.Get("orderBy"),
		BusinessID:       values.Get("business_id"),
		Name:             values.Get("name"),
		Desc:             values.Get("description"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
	}

	return filter, nil
}

func parseFilter(qp queryParams) (business.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter business.QueryFilter

	if qp.BusinessID != "" {
		id, err := uuid.Parse(qp.BusinessID)
		switch err {
		case nil:
			filter.WithBusinessID(id)
		default:
			fieldErrors.Add("business_id", err)
		}
	}

	if qp.Name != "" {
		filter.WithName(qp.Name)
	}

	if qp.Desc != "" {
		filter.WithDesc(qp.Desc)
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
			filter.WithEndCreatedDate(t)
		default:
			fieldErrors.Add(qp.EndCreatedDate, err)
		}
	}

	if err := filter.Validate(); err != nil {
		fieldErrors.Add("filter validation", err)
	}

	if fieldErrors != nil {
		return business.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
