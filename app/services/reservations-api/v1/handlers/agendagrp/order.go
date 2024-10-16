package agendagrp

import (
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/foundation/validate"
)

func parseGeneralAgendaOrder(r *http.Request) (order.By, error) {
	const (
		orderByID         = "id"
		orderByBusinessID = "business_id"
	)

	var orderByFields = map[string]string{
		orderByID:         agenda.OrderByID,
		orderByBusinessID: agenda.OrderByBusinessID,
	}

	orderBy, err := order.Parse(r, order.NewBy(orderByID, order.ASC))
	if err != nil {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	orderBy.Field = orderByFields[orderBy.Field]
	return orderBy, nil
}

func parseDailyAgendaOrder(r *http.Request) (order.By, error) {
	const (
		orderByID         = "id"
		orderByBusinessID = "business_id"
	)

	var orderByFields = map[string]string{
		orderByID:         agenda.OrderByID,
		orderByBusinessID: agenda.OrderByBusinessID,
	}

	orderBy, err := order.Parse(r, order.NewBy(orderByID, order.ASC))
	if err != nil {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	orderBy.Field = orderByFields[orderBy.Field]
	return orderBy, nil
}
