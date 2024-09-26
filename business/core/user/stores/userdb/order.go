package userdb

import (
	"fmt"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/order"
)

// orderByFields is a mappings of order field names in core layer
// to database field names
var orderByFields = map[string]string{
	user.OrderByID:          "user_id",
	user.OrderByName:        "name",
	user.OrderByEmail:       "email",
	user.OrderByRoles:       "roles",
	user.OrderByPhoneNumber: "phone_no",
	user.OrderByEnabled:     "enabled",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exists", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
