package agendadb

import (
	"bytes"
	"strings"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
)

func (s *Store) applyFilterGeneralAgenda(filter agenda.GAQueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}
	if filter.BusinesesID != nil {
		data["business_id"] = *filter.BusinesesID
		wc = append(wc, "business_id = :business_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.Write([]byte(strings.Join(wc, " AND ")))
	}
}

func (s *Store) applyFilterDailyAgenda(filter agenda.DAQueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Date != nil {
		data["date"] = *filter.Date
		wc = append(wc, "applicable_date = :date")
	}

	if filter.From != nil {
		data["from"] = *filter.From
		wc = append(wc, "applicable_date >= :from")
	}

	if filter.To != nil {
		data["to"] = *filter.To
		wc = append(wc, "applicable_date <= :to")
	}

	if filter.Days != nil {
		data["now"] = time.Now().UTC().Format(time.DateOnly)
		data["then"] = time.Now().UTC().AddDate(0, 0, *filter.Days).Format(time.DateOnly)
		wc = append(wc, "applicable_date >= :now AND applicable_date < :then")
	}

	if filter.ID == nil && filter.Date == nil && filter.From == nil && filter.To == nil && filter.Days == nil {
		data["now"] = time.Now().UTC().Format(time.DateOnly)
		wc = append(wc, "applicable_date >= :now")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.Write([]byte(strings.Join(wc, " AND ")))
	}
}
