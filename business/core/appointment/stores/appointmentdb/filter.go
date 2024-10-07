package appointmentdb

import (
	"bytes"
	"strings"

	"github.com/ameghdadian/service/business/core/appointment"
)

func (s *Store) applyFilter(filter appointment.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["appointment_id"] = *filter.ID
		wc = append(wc, "appointment_id = :appointment_id")
	}

	if filter.BusinessID != nil {
		data["business_id"] = *filter.BusinessID
		wc = append(wc, "business_id = :business_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.Status != nil {
		data["status"] = toDBStatus(*filter.Status)
		wc = append(wc, "status = :status")
	}

	if filter.ScheduledOn != nil {
		data["scheduled_on"] = *filter.ScheduledOn
		wc = append(wc, "scheduled_on = :scheduled_on")
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
		buf.Write([]byte(strings.Join(wc, " AND ")))
	}
}
