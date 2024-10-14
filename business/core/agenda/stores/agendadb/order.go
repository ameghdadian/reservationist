package agendadb

import (
	"fmt"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/data/order"
)

var orderByFields = map[string]string{
	agenda.OrderByID: "id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("filed %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
