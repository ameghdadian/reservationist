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

	if filter.BusinessID != nil {
		data["bsnID"] = *filter.BusinessID
		wc = append(wc, "business_id = :bsnID")
	}

	if filter.Date != nil {
		d := *filter.Date
		data["date_start"] = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		data["date_end"] = time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, d.Location())
		wc = append(wc, "opens_at >= :date_start AND opens_at <= :date_end")
	}

	if filter.From != nil {
		data["from"] = *filter.From
		wc = append(wc, "opens_at >= :from")
	}

	if filter.To != nil {
		data["to"] = *filter.To
		wc = append(wc, "opens_at <= :to")
	}

	if filter.Days != nil {
		data["now"] = time.Now().UTC()
		data["then"] = time.Now().UTC().AddDate(0, 0, *filter.Days)
		wc = append(wc, "opens_at >= :now AND opens_at < :then")
	}

	if filter.ID == nil &&
		filter.Date == nil &&
		filter.From == nil &&
		filter.To == nil &&
		filter.Days == nil {
		data["now"] = time.Now().UTC()
		wc = append(wc, "opens_at >= :now")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.Write([]byte(strings.Join(wc, " AND ")))
	}
}
