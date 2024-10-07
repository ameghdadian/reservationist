package appointmentdb

import (
	"time"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/google/uuid"
)

var toDBStatuses = map[appointment.Status]int16{
	appointment.Cancelled: 0,
	appointment.Scheduled: 1,
}

var toCoreStatuses = map[int16]appointment.Status{
	0: appointment.Cancelled,
	1: appointment.Scheduled,
}

// toDBStatus converts status string value to db smallint
func toDBStatus(st appointment.Status) int16 {
	return toDBStatuses[st]
}

// toCoreStatus converts status string value to db smallint
func toCoreStatus(val int16) appointment.Status {
	return toCoreStatuses[val]
}

type dbAppointment struct {
	ID          uuid.UUID `db:"appointment_id"`
	BusinessID  uuid.UUID `db:"business_id"`
	UserID      uuid.UUID `db:"user_id"`
	Status      int16     `db:"status"`
	ScheduledOn time.Time `db:"scheduled_on"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toDBAppointment(apt appointment.Appointment) dbAppointment {
	return dbAppointment{
		ID:          apt.ID,
		BusinessID:  apt.BusinessID,
		UserID:      apt.UserID,
		Status:      toDBStatus(apt.Status),
		ScheduledOn: apt.ScheduledOn.UTC(),
		DateCreated: apt.DateCreated.UTC(),
		DateUpdated: apt.DateUpdated.UTC(),
	}
}

func toCoreAppointment(dbApt dbAppointment) appointment.Appointment {
	return appointment.Appointment{
		ID:          dbApt.ID,
		BusinessID:  dbApt.BusinessID,
		UserID:      dbApt.UserID,
		Status:      toCoreStatus(dbApt.Status),
		ScheduledOn: dbApt.ScheduledOn.In(time.Local),
		DateCreated: dbApt.DateCreated.In(time.Local),
		DateUpdated: dbApt.DateUpdated.In(time.Local),
	}
}

func toCoreAppointmentSlice(dbApts []dbAppointment) []appointment.Appointment {
	apts := make([]appointment.Appointment, len(dbApts))
	for i, dbApt := range dbApts {
		apts[i] = toCoreAppointment(dbApt)
	}

	return apts
}
