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
		data["id"] = *filter.BusinesesID
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
		data["days"] = *filter.Days
		data["now"] = time.Now().UTC()
		wc = append(wc, "applicable_date >= :now AND applicable_date < :now + INTERVAL ':days days'")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.Write([]byte(strings.Join(wc, " AND ")))
	}
}
