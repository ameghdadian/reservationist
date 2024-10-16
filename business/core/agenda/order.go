package agenda

import "github.com/ameghdadian/service/business/data/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID         = "id"
	OrderByBusinessID = "business_id"
)
