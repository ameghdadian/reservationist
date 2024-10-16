package agendagrp

import (
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/google/uuid"
)

type AppGeneralAgenda struct {
	ID          string `json:"id"`
	BusinessID  string `json:"business_id"`
	OpensAt     int    `json:"open_at"`
	ClosedAt    int    `json:"closed_at"`
	Interval    int    `json:"interval"`
	WorkingDays []int  `json:"working_days"`
	TZ          string `json:"timezone"`
	DateCreated string `json:"date_created"`
	DateUpdated string `json:"date_updated"`
}

func toAppGeneralAgenda(agd agenda.GeneralAgenda) AppGeneralAgenda {
	days := make([]int, len(agd.WorkingDays))
	for i, d := range agd.WorkingDays {
		days[i] = int(d.DayOfWeedk())
	}

	opn := int(agd.OpensAt.Hour()*3600 + agd.OpensAt.Minute()*60 + agd.OpensAt.Second())
	cld := int(agd.ClosedAt.Hour()*3600 + agd.ClosedAt.Minute()*60 + agd.ClosedAt.Second())

	return AppGeneralAgenda{
		ID:          agd.ID.String(),
		BusinessID:  agd.BusinessID.String(),
		OpensAt:     opn,
		ClosedAt:    cld,
		Interval:    int(agd.Interval),
		WorkingDays: days,
		TZ:          time.Local.String(),
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
	OpensAt     int    `json:"opens_at" validate:"required,gt=0,lte=86400,required_with=ClosedAt"`
	ClosedAt    int    `json:"closed_at" validate:"required,gt=0,lte=86400,gtfield=OpensAt,required_with=OpensAt"`
	Interval    int    `json:"interval" validate:"required,gt=0,lte=86400"`
	WorkingDays []int  `json:"working_days" validate:"required,max=7,dive,gte=0,lte=6"`
	TZ          string `json:"timezone" validate:"required,timezone"`
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

	loc, _ := time.LoadLocation(app.TZ)
	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	return agenda.NewGeneralAgenda{
		BusinessID:  bsnID,
		OpensAt:     midnight.Add(time.Duration(app.OpensAt) * time.Second),
		ClosedAt:    midnight.Add(time.Duration(app.ClosedAt) * time.Second),
		Interval:    time.Duration(app.Interval) * time.Second,
		WorkingDays: days,
	}, nil
}

// ---------------------------------------------------------------------------------

type AppUpdateGeneralAgenda struct {
	OpensAt     *int   `json:"opens_at" validate:"required,gt=0,lte=86400,required_with=ClosedAt"`
	ClosedAt    *int   `json:"closed_at" validate:"required,gt=0,lte=86400,gtfield=OpensAt,required_with=OpensAt"`
	Interval    *int   `json:"interval" validate:"required,gt=0,lte=86400"`
	WorkingDays []int  `json:"working_days" validate:"required,max=7,dive,gte=0,lte=6"`
	TZ          string `json:"timezone" validate:"required,timezone"`
}

func (app AppUpdateGeneralAgenda) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}

func toCoreUpdateGeneralAgenda(app AppUpdateGeneralAgenda) (agenda.UpdateGeneralAgenda, error) {
	loc, _ := time.LoadLocation(app.TZ)
	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	var opn *time.Time
	if app.OpensAt != nil {
		o := midnight.Add(time.Duration(*app.OpensAt) * time.Second)
		opn = TimePointer(o)
	}
	var cld *time.Time
	if app.ClosedAt != nil {
		c := midnight.Add(time.Duration(*app.ClosedAt) * time.Second)
		cld = TimePointer(c)
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
		Interval:    DurationPointer(time.Duration(*app.Interval) * time.Second),
		WorkingDays: days,
	}, nil

}

// =================================================================================
// =================================================================================

type AppDailyAgenda struct {
	ID           string `json:"id"`
	BusinessID   string `json:"business_id"`
	OpensAt      int    `json:"opens_at"`
	ClosedAt     int    `json:"closed_at"`
	Interval     int    `json:"interval"`
	Date         string `json:"date"`
	Availability bool   `json:"availability"`
	TZ           string `json:"timezone"`
	DateCreated  string `json:"date_created"`
	DateUpdated  string `json:"date_updated"`
}

func toAppDailyAgenda(agd agenda.DailyAgenda) AppDailyAgenda {
	opn := int(agd.OpensAt.Hour()*3600 + agd.OpensAt.Minute()*60 + agd.OpensAt.Second())
	cld := int(agd.ClosedAt.Hour()*3600 + agd.ClosedAt.Minute()*60 + agd.ClosedAt.Second())

	return AppDailyAgenda{
		ID:           agd.ID.String(),
		BusinessID:   agd.BusinessID.String(),
		OpensAt:      opn,
		ClosedAt:     cld,
		Interval:     int(time.Duration(agd.Interval)),
		Date:         agd.Date.Format(time.DateOnly),
		Availability: agd.Availability,
		TZ:           time.Local.String(),
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
	OpensAt      int    `json:"opens_at" validate:"gt=0,lte=86400,required_with=ClosedAt"`
	ClosedAt     int    `json:"closed_at" validate:"gt=0,lte=86400,gtfield=OpensAt,required_with=OpensAt"`
	Interval     int    `json:"interval" validate:"gt=0,lte=86400"`
	Date         string `json:"date" validate:"required"`
	Availability bool   `json:"availability" validate:"required"`
	TZ           string `json:"timezone" validate:"required,timezone"`
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

	loc, _ := time.LoadLocation(app.TZ)
	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	return agenda.NewDailyAgenda{
		BusinessID:   bsnID,
		OpensAt:      midnight.Add(time.Duration(app.OpensAt) * time.Second),
		ClosedAt:     midnight.Add(time.Duration(app.ClosedAt) * time.Second),
		Interval:     time.Duration(app.Interval) * time.Second,
		Date:         date,
		Availability: app.Availability,
	}, nil
}

// ---------------------------------------------------------------------------------

type AppUpdateDailyAgenda struct {
	OpensAt      *int    `json:"opens_at" validate:"gt=0,lte=86400,required_with=ClosedAt"`
	ClosedAt     *int    `json:"closed_at" validate:"gt=0,lte=86400,gtfield=OpensAt,required_with=OpensAt"`
	Interval     *int    `json:"interval" validate:"gt=0,lte=86400"`
	Date         *string `json:"date" validate:"required"`
	Availability *bool   `json:"availability" validate:"required"`
	TZ           *string `json:"timezone" validate:"timezone,required_with=OpensAt"`
}

func toCoreUpdateDailyAgenda(app AppUpdateDailyAgenda) (agenda.UpdateDailyAgenda, error) {
	var loc *time.Location
	var now, midnight time.Time
	if app.TZ != nil {
		loc, _ = time.LoadLocation(*app.TZ)
		now = time.Now().In(loc)
		midnight = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	var opn *time.Time
	if app.OpensAt != nil {
		o := midnight.Add(time.Duration(*app.OpensAt) * time.Second)
		opn = TimePointer(o)
	}
	var cld *time.Time
	if app.ClosedAt != nil {
		c := midnight.Add(time.Duration(*app.ClosedAt) * time.Second)
		cld = TimePointer(c)
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
		Interval:     DurationPointer(time.Duration(*app.Interval) * time.Second),
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

func DurationPointer(d time.Duration) *time.Duration {
	return &d
}
