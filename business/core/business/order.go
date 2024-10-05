package business

import "github.com/ameghdadian/service/business/data/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID      = "business_id"
	OrderByOwnerID = "owner_id"
	OrderByName    = "name"
	OrderByDesc    = "desc"
)
