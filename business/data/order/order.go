package order

import (
	"fmt"
	"strings"
)

const (
	ASC  = "ASC"
	DESC = "DESC"
)

var directions = map[string]string{
	ASC:  "ASC",
	DESC: "DESC",
}

// ===================================================

type By struct {
	Field     string
	Direction string
}

func NewBy(field string, direction string) By {
	return By{
		Field:     field,
		Direction: direction,
	}
}

// Parse construct a order.By value.
// Order of appearance in the URL query param is: "?orderBy=field,direction"
func Parse(fieldMappings map[string]string, orderBy string, defaultOrderBy By) (By, error) {
	if orderBy == "" {
		return defaultOrderBy, nil
	}

	orderParts := strings.Split(orderBy, ",")

	orgFieldName := strings.TrimSpace(orderParts[0])
	fieldName, exists := fieldMappings[orgFieldName]
	if !exists {
		return By{}, fmt.Errorf("unknown order: %s", orgFieldName)
	}

	switch len(orderParts) {
	case 1:
		return NewBy(fieldName, ASC), nil

	case 2:
		direction := strings.TrimSpace(orderParts[1])
		if _, exists := directions[direction]; !exists {
			return By{}, fmt.Errorf("unknown direction: %s", direction)
		}

		return NewBy(fieldName, direction), nil

	default:
		return By{}, fmt.Errorf("unknown order: %s", orderBy)
	}
}
