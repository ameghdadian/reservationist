package agendadb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/data/dbsql/pgx/dbarray"
	"github.com/google/uuid"
)

type dbGeneralAgenda struct {
	ID          uuid.UUID     `db:"id"`
	BusinessID  uuid.UUID     `db:"business_id"`
	OpensAt     time.Time     `db:"opens_at"`
	ClosedAt    time.Time     `db:"closed_at"`
	Interval    int           `db:"interval"`
	WorkingDays dbarray.Int32 `db:"working_days"`
	DateCreated time.Time     `db:"date_created"`
	DateUpdated time.Time     `db:"date_updated"`
}

func toDBGeneralAgenda(gAgd agenda.GeneralAgenda) dbGeneralAgenda {
	days := make([]int32, len(gAgd.WorkingDays))
	for i, d := range gAgd.WorkingDays {
		days[i] = int32(d.DayOfWeedk())
	}

	return dbGeneralAgenda{
		ID:          gAgd.ID,
		BusinessID:  gAgd.BusinessID,
		OpensAt:     gAgd.OpensAt.UTC(),
		ClosedAt:    gAgd.ClosedAt.UTC(),
		Interval:    int(gAgd.Interval.Seconds()),
		WorkingDays: days,
		DateCreated: gAgd.DateCreated.UTC(),
		DateUpdated: gAgd.DateUpdated.UTC(),
	}
}

func toCoreGeneralAgenda(dbAgd dbGeneralAgenda) (agenda.GeneralAgenda, error) {
	days := make([]agenda.Day, len(dbAgd.WorkingDays))
	for i, dd := range dbAgd.WorkingDays {
		var err error
		days[i], err = agenda.ParseDay(uint(dd))
		if err != nil {
			return agenda.GeneralAgenda{}, fmt.Errorf("parse day: %w", err)
		}
	}

	return agenda.GeneralAgenda{
		ID:          dbAgd.ID,
		BusinessID:  dbAgd.BusinessID,
		OpensAt:     dbAgd.OpensAt.In(time.Local),
		ClosedAt:    dbAgd.ClosedAt.In(time.Local),
		Interval:    time.Duration(dbAgd.Interval) * time.Second,
		WorkingDays: days,
		DateCreated: dbAgd.DateCreated.In(time.Local),
		DateUpdated: dbAgd.DateUpdated.In(time.Local),
	}, nil
}

func toCoreGeneralAgendaSlice(dbgAgds []dbGeneralAgenda) ([]agenda.GeneralAgenda, error) {
	agds := make([]agenda.GeneralAgenda, len(dbgAgds))

	for i, agd := range dbgAgds {
		var err error
		agds[i], err = toCoreGeneralAgenda(agd)
		if err != nil {
			return nil, err
		}
	}

	return agds, nil
}

// ---------------------------------------------------------------------------------

type dbDailyAgenda struct {
	ID           uuid.UUID     `db:"id"`
	BusinessID   uuid.UUID     `db:"business_id"`
	OpensAt      sql.NullTime  `db:"opens_at"`
	ClosedAt     sql.NullTime  `db:"closed_at"`
	Interval     sql.NullInt64 `db:"interval"`
	Date         string        `db:"applicable_date"`
	Availability bool          `db:"availability"`
	DateCreated  time.Time     `db:"date_created"`
	DateUpdated  time.Time     `db:"date_updated"`
}

func toDBDailyAgenda(gAgd agenda.DailyAgenda) dbDailyAgenda {
	return dbDailyAgenda{
		ID:         gAgd.ID,
		BusinessID: gAgd.BusinessID,
		OpensAt: sql.NullTime{
			Time:  gAgd.OpensAt.UTC(),
			Valid: !gAgd.OpensAt.UTC().IsZero(),
		},
		ClosedAt: sql.NullTime{
			Time:  gAgd.ClosedAt.UTC(),
			Valid: !gAgd.ClosedAt.UTC().IsZero(),
		},
		Interval: sql.NullInt64{
			Int64: int64(gAgd.Interval.Seconds()),
			Valid: gAgd.Interval.Seconds() >= 0,
		},
		Date:         gAgd.Date.UTC().Format(time.DateOnly),
		Availability: true,
		DateCreated:  gAgd.DateCreated.UTC(),
		DateUpdated:  gAgd.DateUpdated.UTC(),
	}
}

func toCoreDailyAgenda(dbAgd dbDailyAgenda) (agenda.DailyAgenda, error) {
	date, err := time.Parse(time.RFC3339, dbAgd.Date)
	if err != nil {
		return agenda.DailyAgenda{}, err
	}

	return agenda.DailyAgenda{
		ID:           dbAgd.ID,
		BusinessID:   dbAgd.BusinessID,
		OpensAt:      dbAgd.OpensAt.Time.In(time.Local),
		ClosedAt:     dbAgd.ClosedAt.Time.In(time.Local),
		Interval:     time.Duration(dbAgd.Interval.Int64) * time.Second,
		Date:         date.In(time.Local),
		Availability: dbAgd.Availability,
		DateCreated:  dbAgd.DateCreated.In(time.Local),
		DateUpdated:  dbAgd.DateUpdated.In(time.Local),
	}, nil
}

func toCoreDailyAgendaSlice(dbgAgds []dbDailyAgenda) ([]agenda.DailyAgenda, error) {
	agds := make([]agenda.DailyAgenda, len(dbgAgds))

	for i, agd := range dbgAgds {
		var err error
		agds[i], err = toCoreDailyAgenda(agd)
		if err != nil {
			return nil, err
		}
	}

	return agds, nil
}
