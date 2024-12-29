package appointmentgrp

import (
	"github.com/ameghdadian/service/business/core/appointment"
)

var orderByFields = map[string]string{
	"appointment_id": appointment.OrderByID,
	"business_id":    appointment.OrderByBusinessID,
	"user_id":        appointment.OrderByUserID,
	"status":         appointment.OrderByStatus,
	"scheduled_date": appointment.OrderByScheduledDate,
}
