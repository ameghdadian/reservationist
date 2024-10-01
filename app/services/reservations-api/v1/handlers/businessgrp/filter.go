package businessgrp

import (
	"net/http"
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(r *http.Request) (business.QueryFilter, error) {
	const (
		filterByBusinessID       = "business_id"
		filterByName             = "name"
		filterByDesc             = "description"
		filterByStartCreatedDate = "start_created_date"
		filterByEndCreatedDate   = "end_created_date"
	)

	values := r.URL.Query()

	var filter business.QueryFilter

	if businessID := values.Get(filterByBusinessID); businessID != "" {
		id, err := uuid.Parse(businessID)
		if err != nil {
			return business.QueryFilter{}, validate.NewFieldsError(filterByBusinessID, err)
		}
		filter.WithBusinessID(id)
	}

	if name := values.Get(filterByName); name != "" {
		filter.WithName(name)
	}

	if desc := values.Get(filterByDesc); desc != "" {
		filter.WithDesc(desc)
	}

	if createdDate := values.Get(filterByStartCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return business.QueryFilter{}, validate.NewFieldsError(filterByStartCreatedDate, err)
		}
		filter.WithStartCreatedDate(t)
	}

	if createdDate := values.Get(filterByEndCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return business.QueryFilter{}, validate.NewFieldsError(filterByEndCreatedDate, err)
		}
		filter.WithEndCreatedDate(t)
	}

	if err := filter.Validate(); err != nil {
		return business.QueryFilter{}, err
	}

	return filter, nil
}
