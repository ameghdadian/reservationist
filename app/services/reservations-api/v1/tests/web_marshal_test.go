package tests

import (
	"time"

	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/agendagrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/appointmentgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/businessgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/usergrp"
	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
)

func toAppUser(usr user.User) usergrp.AppUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return usergrp.AppUser{
		ID:           usr.ID.String(),
		Name:         usr.Name,
		Email:        usr.Email.Address,
		Roles:        roles,
		PasswordHash: nil, // This field is not marshalled.
		PhoneNo:      usr.PhoneNo.Number(),
		Enabled:      usr.Enabled,
		DateCreated:  usr.DateCreated.Format(time.RFC3339),
		DateUpdated:  usr.DateUpdated.Format(time.RFC3339),
	}
}

func toAppUsers(users []user.User) []usergrp.AppUser {
	items := make([]usergrp.AppUser, len(users))
	for i, usr := range users {
		items[i] = toAppUser(usr)
	}

	return items
}

func toAppUserPtr(usr user.User) *usergrp.AppUser {
	appUsr := toAppUser(usr)
	return &appUsr
}

// ----------------------------------------------------------

func toAppBusiness(b business.Business) businessgrp.AppBusiness {
	return businessgrp.AppBusiness{
		ID:          b.ID.String(),
		OwnerID:     b.OwnerID.String(),
		Name:        b.Name,
		Description: b.Desc,
		DateCreated: b.DateCreated.Format(time.RFC3339),
		DateUpdated: b.DateUpdated.Format(time.RFC3339),
	}
}

func toAppBusinesses(bsns []business.Business) []businessgrp.AppBusiness {
	items := make([]businessgrp.AppBusiness, len(bsns))
	for i, b := range bsns {
		items[i] = toAppBusiness(b)
	}

	return items
}

func toAppBusinessPtr(b business.Business) *businessgrp.AppBusiness {
	appBsn := toAppBusiness(b)
	return &appBsn
}

// ----------------------------------------------------------

func toAppAppointment(apt appointment.Appointment) appointmentgrp.AppAppointment {
	return appointmentgrp.AppAppointment{
		ID:          apt.ID.String(),
		BusinessID:  apt.BusinessID.String(),
		UserID:      apt.UserID.String(),
		Status:      apt.Status.Status(),
		ScheduledOn: apt.ScheduledOn.Format(time.RFC3339),
		DateCreated: apt.DateCreated.Format(time.RFC3339),
		DateUpdated: apt.DateUpdated.Format(time.RFC3339),
	}
}

func toAppAppointments(apts []appointment.Appointment) []appointmentgrp.AppAppointment {
	apps := make([]appointmentgrp.AppAppointment, len(apts))
	for i, apt := range apts {
		apps[i] = toAppAppointment(apt)
	}

	return apps
}

func toAppAppointmentPtr(a appointment.Appointment) *appointmentgrp.AppAppointment {
	appApt := toAppAppointment(a)
	return &appApt
}

// ----------------------------------------------------------

func toAppGeneralAgenda(agd agenda.GeneralAgenda) agendagrp.AppGeneralAgenda {
	days := make([]int, len(agd.WorkingDays))
	for i, d := range agd.WorkingDays {
		days[i] = int(d.DayOfWeedk())
	}

	return agendagrp.AppGeneralAgenda{
		ID:          agd.ID.String(),
		BusinessID:  agd.BusinessID.String(),
		OpensAt:     agd.OpensAt.Format(time.RFC3339),
		ClosedAt:    agd.ClosedAt.Format(time.RFC3339),
		Interval:    int(agd.Interval),
		WorkingDays: days,
		DateCreated: agd.DateCreated.Format(time.RFC3339),
		DateUpdated: agd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppGeneralAgendas(agds []agenda.GeneralAgenda) []agendagrp.AppGeneralAgenda {
	coll := make([]agendagrp.AppGeneralAgenda, len(agds))
	for i, a := range agds {
		coll[i] = toAppGeneralAgenda(a)
	}

	return coll
}

func toAppGeneralAgendaPtr(app agenda.GeneralAgenda) *agendagrp.AppGeneralAgenda {
	agd := toAppGeneralAgenda(app)
	return &agd
}

func toAppDailyAgenda(agd agenda.DailyAgenda) agendagrp.AppDailyAgenda {
	return agendagrp.AppDailyAgenda{
		ID:           agd.ID.String(),
		BusinessID:   agd.BusinessID.String(),
		OpensAt:      agd.OpensAt.Format(time.RFC3339),
		ClosedAt:     agd.ClosedAt.Format(time.RFC3339),
		Interval:     int(time.Duration(agd.Interval)),
		Date:         agd.Date.Format(time.DateOnly),
		Availability: agd.Availability,
		DateCreated:  agd.DateCreated.Format(time.RFC3339),
		DateUpdated:  agd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppDailyAgendas(agds []agenda.DailyAgenda) []agendagrp.AppDailyAgenda {
	col := make([]agendagrp.AppDailyAgenda, len(agds))
	for i, a := range agds {
		col[i] = toAppDailyAgenda(a)
	}

	return col
}

func toAppDailyAgendaPtr(app agenda.DailyAgenda) *agendagrp.AppDailyAgenda {
	agd := toAppDailyAgenda(app)
	return &agd
}
