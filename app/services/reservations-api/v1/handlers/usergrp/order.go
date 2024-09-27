package usergrp

import (
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/foundation/validate"
)

func parseOrder(r *http.Request) (order.By, error) {
	const (
		orderByUserID      = "user_id"
		orderByEmail       = "email"
		orderByName        = "name"
		orderByPhoneNumber = "phone_number"
		orderByRoles       = "roles"
		orderByEnabled     = "enabled"
	)

	var orderByFields = map[string]string{
		orderByUserID:      user.OrderByID,
		orderByEmail:       user.OrderByEmail,
		orderByName:        user.OrderByName,
		orderByPhoneNumber: user.OrderByPhoneNumber,
		orderByRoles:       user.OrderByRoles,
		orderByEnabled:     user.OrderByEnabled,
	}

	orderBy, err := order.Parse(r, order.NewBy(orderByUserID, order.ASC))
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	orderBy.Field = orderByFields[orderBy.Field]

	return orderBy, nil
}
