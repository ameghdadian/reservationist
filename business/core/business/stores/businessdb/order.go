package businessdb

import (
	"fmt"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/data/order"
)

var orderByFields = map[string]string{
	business.OrderByID:      "business_id",
	business.OrderByOwnerID: "owner_id",
	business.OrderByName:    "name",
	business.OrderByDesc:    "desc",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exists", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
