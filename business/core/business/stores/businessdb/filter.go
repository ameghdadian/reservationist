package businessdb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ameghdadian/service/business/core/business"
)

func (s *Store) applyFilter(filter business.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["business_id"] = *filter.ID
		wc = append(wc, "business_id = :business_id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Desc != nil {
		data["desc"] = fmt.Sprintf("%%%s%%", *filter.Desc)
		wc = append(wc, "description LIKE :desc")
	}

	if filter.StartCreatedDate != nil {
		data["start_date_created"] = *filter.StartCreatedDate
		wc = append(wc, "date_created >= :start_date_created")
	}

	if filter.EndCreatedDate != nil {
		data["end_date_created"] = *filter.EndCreatedDate
		wc = append(wc, "date_created <= :end_date_created")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
