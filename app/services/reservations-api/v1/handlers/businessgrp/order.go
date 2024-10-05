package businessgrp

import (
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/foundation/validate"
)

func parseOrder(r *http.Request) (order.By, error) {
	const (
		orderByID      = "business_id"
		orderByOwnerID = "owner_id"
		orderByName    = "name"
		orderByDesc    = "description"
	)

	var orderByFields = map[string]string{
		orderByID:      business.OrderByID,
		orderByOwnerID: business.OrderByOwnerID,
		orderByName:    business.OrderByName,
		orderByDesc:    business.OrderByDesc,
	}

	orderBy, err := order.Parse(r, order.NewBy(orderByID, order.ASC))
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	orderBy.Field = orderByFields[orderBy.Field]

	return orderBy, nil
}
