package agendagrp

import (
	"errors"
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/google/uuid"
)

type AppGeneralAgenda struct {
	ID          string `json:"id"`
	BusinessID  string `json:"business_id"`
	OpensAt     string `json:"open_at"`
	ClosedAt    string `json:"closed_at"`
	Interval    int    `json:"interval"`
	WorkingDays []int  `json:"working_days"`
	DateCreated string `json:"date_created"`
	DateUpdated string `json:"date_updated"`
}

func toAppGeneralAgenda(agd agenda.GeneralAgenda) AppGeneralAgenda {
	days := make([]int, len(agd.WorkingDays))
	for i, d := range agd.WorkingDays {
		days[i] = int(d.DayOfWeedk())
	}

	return AppGeneralAgenda{
		ID:          agd.ID.String(),
		BusinessID:  agd.BusinessID.String(),
		OpensAt:     agd.OpensAt.Format(time.RFC3339),
		ClosedAt:    agd.ClosedAt.Format(time.RFC3339),
		Interval:    agd.Interval,
		WorkingDays: days,
		DateCreated: agd.DateCreated.Format(time.RFC3339),
		DateUpdated: agd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppGeneralAgendaSlice(agds []agenda.GeneralAgenda) []AppGeneralAgenda {
	coll := make([]AppGeneralAgenda, len(agds))
	for i, a := range agds {
		coll[i] = toAppGeneralAgenda(a)
	}

	return coll
}

// ---------------------------------------------------------------------------------

type AppNewGeneralAgenda struct {
	BusinessID  string `json:"business_id" validate:"required,uuid"`
	OpensAt     string `json:"opens_at" validate:"required,required_with=ClosedAt"`
	ClosedAt    string `json:"closed_at" validate:"required,required_with=OpensAt"`
	Interval    int    `json:"interval" validate:"required,gt=0,lte=86400"`
	WorkingDays []int  `json:"working_days" validate:"required,max=7,dive,gte=0,lte=6"`
}

func (app AppNewGeneralAgenda) Validate() error {
	if err := validate.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toCoreNewGeneralAgenda(app AppNewGeneralAgenda) (agenda.NewGeneralAgenda, error) {
	bsnID, err := uuid.Parse(app.BusinessID)
	if err != nil {
		return agenda.NewGeneralAgenda{}, fmt.Errorf("parsing business id: %w", err)
	}

	days := make([]agenda.Day, len(app.WorkingDays))
	for i, d := range app.WorkingDays {
		day, err := agenda.ParseDay(uint(d))
		if err != nil {
			return agenda.NewGeneralAgenda{}, fmt.Errorf("parsing day: %w", err)
		}
		days[i] = day
	}

	opn, err := time.Parse(time.RFC3339, app.OpensAt)
	if err != nil {
		return agenda.NewGeneralAgenda{}, fmt.Errorf("parsing opens at: %w", err)
	}
	cld, err := time.Parse(time.RFC3339, app.ClosedAt)
	if err != nil {
		return agenda.NewGeneralAgenda{}, fmt.Errorf("parsing closed at: %w", err)
	}

	if opn.Format(time.DateOnly) != cld.Format(time.DateOnly) {
		return agenda.NewGeneralAgenda{}, errors.New("opening and closing hour can not be in two separate days")
	}

	if cld.Before(opn) {
		return agenda.NewGeneralAgenda{}, errors.New("Closed at time should be after Opens at time.")
	}

	return agenda.NewGeneralAgenda{
		BusinessID:  bsnID,
		OpensAt:     opn,
		ClosedAt:    cld,
		Interval:    app.Interval,
		WorkingDays: days,
	}, nil
}

// ---------------------------------------------------------------------------------

type AppUpdateGeneralAgenda struct {
	OpensAt     *string `json:"opens_at" validate:"omitempty,required_with=ClosedAt"`
	ClosedAt    *string `json:"closed_at" validate:"omitempty,required_with=OpensAt"`
	Interval    *int    `json:"interval" validate:"omitempty,gt=0,lte=86400"`
	WorkingDays []int   `json:"working_days" validate:"omitempty,max=7,dive,gte=0,lte=6"`
}

func (app AppUpdateGeneralAgenda) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}

func toCoreUpdateGeneralAgenda(app AppUpdateGeneralAgenda) (agenda.UpdateGeneralAgenda, error) {
	var opn *time.Time
	if app.OpensAt != nil {
		o, err := time.Parse(time.RFC3339, *app.OpensAt)
		if err != nil {
			return agenda.UpdateGeneralAgenda{}, fmt.Errorf("parsing opens at: %w", err)
		}
		opn = TimePointer(o)
	}
	var cld *time.Time
	if app.ClosedAt != nil {
		c, err := time.Parse(time.RFC3339, *app.ClosedAt)
		if err != nil {
			return agenda.UpdateGeneralAgenda{}, fmt.Errorf("parsing closed at: %w", err)
		}
		cld = TimePointer(c)
	}

	if opn.Format(time.DateOnly) != cld.Format(time.DateOnly) {
		return agenda.UpdateGeneralAgenda{}, errors.New("opening and closing hour can not be in two separate days")
	}

	var days []agenda.Day
	if app.WorkingDays != nil {
		for _, d := range app.WorkingDays {
			day, err := agenda.ParseDay(uint(d))
			if err != nil {
				return agenda.UpdateGeneralAgenda{}, fmt.Errorf("parsing day: %w", err)
			}
			days = append(days, day)
		}
	}

	return agenda.UpdateGeneralAgenda{
		OpensAt:     opn,
		ClosedAt:    cld,
		Interval:    app.Interval,
		WorkingDays: days,
	}, nil

}

// =================================================================================
// =================================================================================

type AppDailyAgenda struct {
	ID           string `json:"id"`
	BusinessID   string `json:"business_id"`
	OpensAt      string `json:"opens_at"`
	ClosedAt     string `json:"closed_at"`
	Interval     int    `json:"interval"`
	Date         string `json:"date"`
	Availability bool   `json:"availability"`
	DateCreated  string `json:"date_created"`
	DateUpdated  string `json:"date_updated"`
}

func toAppDailyAgenda(agd agenda.DailyAgenda) AppDailyAgenda {
	return AppDailyAgenda{
		ID:           agd.ID.String(),
		BusinessID:   agd.BusinessID.String(),
		OpensAt:      agd.OpensAt.Format(time.RFC3339),
		ClosedAt:     agd.ClosedAt.Format(time.RFC3339),
		Interval:     agd.Interval,
		Date:         agd.Date.Format(time.DateOnly),
		Availability: agd.Availability,
		DateCreated:  agd.DateCreated.Format(time.RFC3339),
		DateUpdated:  agd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppDailyAgendaSlice(agds []agenda.DailyAgenda) []AppDailyAgenda {
	col := make([]AppDailyAgenda, len(agds))
	for i, a := range agds {
		col[i] = toAppDailyAgenda(a)
	}

	return col
}

// ---------------------------------------------------------------------------------

type AppNewDailyAgenda struct {
	BusinessID   string `json:"business_id" validate:"required,uuid"`
	OpensAt      string `json:"opens_at" validate:"required_if=Availability true"`
	ClosedAt     string `json:"closed_at" validate:"required_if=Availability true"`
	Interval     int    `json:"interval" validate:"gt=0,lte=86400,required_if=Availability true"`
	Date         string `json:"date" validate:"required"`
	Availability bool   `json:"availability" validate:"required"`
}

func (app AppNewDailyAgenda) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}

func toCoreNewDailyAgenda(app AppNewDailyAgenda) (agenda.NewDailyAgenda, error) {
	bsnID, err := uuid.Parse(app.BusinessID)
	if err != nil {
		return agenda.NewDailyAgenda{}, fmt.Errorf("parsing business id: %w", err)
	}

	date, err := time.Parse(time.DateOnly, app.Date)
	if err != nil {
		return agenda.NewDailyAgenda{}, fmt.Errorf("parsing date: %w", err)
	}

	opn, err := time.Parse(time.RFC3339, app.OpensAt)
	if err != nil {
		return agenda.NewDailyAgenda{}, fmt.Errorf("parsing opens at: %w", err)
	}
	cld, err := time.Parse(time.RFC3339, app.ClosedAt)
	if err != nil {
		return agenda.NewDailyAgenda{}, fmt.Errorf("parsing closed at: %w", err)
	}

	if cld.Before(opn) {
		return agenda.NewDailyAgenda{}, errors.New("closed at time should be after Opens at time")
	}

	if opn.Format(time.DateOnly) != cld.Format(time.DateOnly) {
		return agenda.NewDailyAgenda{}, errors.New("opening and closing hour can not be in two separate days")
	}

	return agenda.NewDailyAgenda{
		BusinessID:   bsnID,
		OpensAt:      opn,
		ClosedAt:     cld,
		Interval:     app.Interval,
		Date:         date,
		Availability: app.Availability,
	}, nil
}

// ---------------------------------------------------------------------------------

type AppUpdateDailyAgenda struct {
	OpensAt      *string `json:"opens_at" validate:"omitempty,required_with=ClosedAt"`
	ClosedAt     *string `json:"closed_at" validate:"omitempty,required_with=OpensAt"`
	Interval     *int    `json:"interval" validate:"omitempty,gt=0,lte=86400"`
	Date         *string `json:"date" validate:"required"`
	Availability *bool   `json:"availability" validate:"required"`
}

func toCoreUpdateDailyAgenda(app AppUpdateDailyAgenda) (agenda.UpdateDailyAgenda, error) {
	var opn *time.Time
	if app.OpensAt != nil {
		o, err := time.Parse(time.RFC3339, *app.OpensAt)
		if err != nil {
			return agenda.UpdateDailyAgenda{}, fmt.Errorf("parsing opens at: %w", err)
		}
		opn = TimePointer(o)
	}
	var cld *time.Time
	if app.ClosedAt != nil {
		c, err := time.Parse(time.RFC3339, *app.ClosedAt)
		if err != nil {
			return agenda.UpdateDailyAgenda{}, fmt.Errorf("parsing closed at: %w", err)
		}
		cld = TimePointer(c)
	}

	if opn.Format(time.DateOnly) != cld.Format(time.DateOnly) {
		return agenda.UpdateDailyAgenda{}, errors.New("opening and closing hour can not be in two separate days")
	}

	var date *time.Time
	if app.Date != nil {
		d, err := time.Parse(time.DateOnly, *app.Date)
		if err != nil {
			return agenda.UpdateDailyAgenda{}, fmt.Errorf("parsing date: %w", err)
		}
		date = TimePointer(d)
	}

	return agenda.UpdateDailyAgenda{
		OpensAt:      opn,
		ClosedAt:     cld,
		Interval:     app.Interval,
		Date:         date,
		Availability: app.Availability,
	}, nil

}

// ---------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------

func TimePointer(t time.Time) *time.Time {
	return &t
}
