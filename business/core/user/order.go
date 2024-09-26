package user

import "github.com/ameghdadian/service/business/data/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// These strings doesn't represent the name of the fields in the storage layer. They'll
// be mapped later with their storage layer counterparts. They should be used by application
// layer.
const (
	OrderByID          = "user_id"
	OrderByName        = "name"
	OrderByEmail       = "email"
	OrderByRoles       = "roles"
	OrderByEnabled     = "enabled"
	OrderByPhoneNumber = "phone_number"
)
