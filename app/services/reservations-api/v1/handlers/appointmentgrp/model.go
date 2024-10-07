package appointmentgrp

import (
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/google/uuid"
)

type AppAppointment struct {
	ID          string `json:"id"`
	BusinessID  string `json:"business_id"`
	UserID      string `json:"user_id"`
	Status      string `json:"string"`
	ScheduledOn string `json:"scheduled_on"`
	DateCreated string `json:"dateCreated"`
	DateUpdated string `json:"dateUpdated"`
}

func toAppAppointment(apt appointment.Appointment) AppAppointment {
	return AppAppointment{
		ID:          apt.ID.String(),
		BusinessID:  apt.BusinessID.String(),
		UserID:      apt.UserID.String(),
		Status:      apt.Status.Status(),
		ScheduledOn: apt.ScheduledOn.Format(time.RFC3339),
		DateCreated: apt.DateCreated.Format(time.RFC3339),
		DateUpdated: apt.DateUpdated.Format(time.RFC3339),
	}
}

func toAppAppointments(apts []appointment.Appointment) []AppAppointment {
	apps := make([]AppAppointment, len(apts))
	for i, apt := range apts {
		apps[i] = toAppAppointment(apt)
	}

	return apps
}

// -------------------------------------------------------------------------------

type AppNewAppointment struct {
	BusinessID  string `json:"business_id" validate:"required,uuid"`
	UserID      string `json:"user_id" validate:"required,uuid"`
	Status      string `json:"status" validate:"required"`
	ScheduledOn string `json:"scheduled_on" validate:"required,datetime"`
}

func (app AppNewAppointment) Validate() error {
	if err := validate.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toCoreNewAppointment(app AppNewAppointment) (appointment.NewAppointment, error) {
	bsnID, err := uuid.Parse(app.BusinessID)
	if err != nil {
		return appointment.NewAppointment{}, fmt.Errorf("parsing business id: %w", err)
	}

	usrID, err := uuid.Parse(app.UserID)
	if err != nil {
		return appointment.NewAppointment{}, fmt.Errorf("parsing user id: %w", err)
	}

	status, err := appointment.ParseStatus(app.Status)
	if err != nil {
		return appointment.NewAppointment{}, fmt.Errorf("parsing status: %w", err)
	}

	sch, err := time.Parse(time.RFC3339, app.ScheduledOn)
	if err != nil {
		return appointment.NewAppointment{}, fmt.Errorf("parsing scheduled on time: %w", err)
	}

	na := appointment.NewAppointment{
		BusinessID:  bsnID,
		UserID:      usrID,
		Status:      status,
		ScheduledOn: sch,
	}

	return na, nil
}

// -------------------------------------------------------------------------------

type AppUpdateAppointment struct {
	Status      *string `json:"status"`
	ScheduledOn *string `json:"scheduled_on" validate:"omitempty,datetime"`
}

// func (app AppUpdateAppointment) Validate() error {

// }

func toCoreUpdateAppointment(app AppUpdateAppointment) (appointment.UpdateAppointment, error) {
	var status appointment.Status
	if app.Status != nil {
		var err error
		status, err = appointment.ParseStatus(*app.Status)
		if err != nil {
			return appointment.UpdateAppointment{}, fmt.Errorf("parsing status: %w", err)
		}
	}

	var t time.Time
	if app.ScheduledOn != nil {
		var err error
		t, err = time.Parse(time.RFC3339, *app.ScheduledOn)
		if err != nil {
			return appointment.UpdateAppointment{}, fmt.Errorf("parsing scheduled on: %w", err)
		}
	}

	apt := appointment.UpdateAppointment{
		Status:      &status,
		ScheduledOn: &t,
	}

	return apt, nil
}
