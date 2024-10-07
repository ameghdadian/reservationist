package appointmentgrp

import (
	"errors"
	"net/http"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/foundation/validate"
)

func parseOrder(r *http.Request) (order.By, error) {
	const (
		orderByID            = "appointment_id"
		orderByBusinessID    = "business_id"
		orderByUserID        = "user_id"
		orderByStatus        = "status"
		orderByScheduledDate = "scheduled_date"
	)

	orderByFields := map[string]string{
		orderByID:            appointment.OrderByID,
		orderByBusinessID:    appointment.OrderByBusinessID,
		orderByUserID:        appointment.OrderByUserID,
		orderByStatus:        appointment.OrderByStatus,
		orderByScheduledDate: appointment.OrderByScheduledDate,
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
