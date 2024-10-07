package appointment

import "github.com/ameghdadian/service/business/data/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID            = "appointment_id"
	OrderByBusinessID    = "business_id"
	OrderByUserID        = "user_id"
	OrderByStatus        = "status"
	OrderByScheduledDate = "scheduled_date"
)
