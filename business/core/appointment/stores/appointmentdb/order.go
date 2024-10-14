package appointmentdb

import (
	"fmt"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/data/order"
)

var orderByFields = map[string]string{
	appointment.OrderByID:            "appointment_id",
	appointment.OrderByBusinessID:    "business_id",
	appointment.OrderByUserID:        "user_id",
	appointment.OrderByStatus:        "status",
	appointment.OrderByScheduledDate: "schedueld_on",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
