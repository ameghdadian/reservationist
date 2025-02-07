package agendadb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ameghdadian/service/business/core/agenda"
	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

func (s *Store) ExecuteUnderTransaction(tx transaction.Transaction) (agenda.Storer, error) {
	ec, err := db.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	s = &Store{
		log: s.log,
		db:  ec,
	}

	return s, nil
}
func (s *Store) CreateGeneralAgenda(ctx context.Context, agd agenda.GeneralAgenda) error {
	const q = `
	INSERT INTO general_agenda
		(id, business_id, opens_at, closed_at, interval, working_days, date_created, date_updated)
	VALUES
		(:id, :business_id, :opens_at, :closed_at, :interval, :working_days, :date_created, :date_updated)
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBGeneralAgenda(agd)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) UpdateGeneralAgenda(ctx context.Context, agd agenda.GeneralAgenda) error {
	const q = `
	UPDATE
		general_agenda
	SET
		"opens_at" = :opens_at,
		"closed_at" = :closed_at,
		"interval" = :interval,
		"working_days" = :working_days,
		"date_updated" = :date_updated
	WHERE
		"id" = :id
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBGeneralAgenda(agd)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) DeleteGeneralAgenda(ctx context.Context, agd agenda.GeneralAgenda) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: agd.ID.String(),
	}

	const q = `
	DELETE FROM
		general_agenda	
	WHERE
		"id" = :id
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexedcontext: %w", err)
	}

	return nil
}

func (s *Store) QueryGeneralAgenda(ctx context.Context, filter agenda.GAQueryFilter, orderBy order.By, page page.Page) ([]agenda.GeneralAgenda, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, business_id, opens_at, closed_at, interval, working_days, date_created, date_updated
	FROM
		general_agenda
	`

	buf := bytes.NewBufferString(q)
	s.applyFilterGeneralAgenda(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY ")

	var dbgAgds []dbGeneralAgenda
	if err := db.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbgAgds); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return nil, fmt.Errorf("namedquerystruct: %w", agenda.ErrNotFound)
		}
		return nil, fmt.Errorf("namedquerystruct: %w", err)
	}

	agds, err := toCoreGeneralAgendaSlice(dbgAgds)
	if err != nil {
		return nil, err
	}

	return agds, nil
}

func (s *Store) QueryGeneralAgendaByBusinessID(ctx context.Context, bsnID uuid.UUID) (agenda.GeneralAgenda, error) {
	data := struct {
		BusinessID string `db:"business_id"`
	}{
		BusinessID: bsnID.String(),
	}

	const q = `
	SELECT 	
		id, business_id, opens_at, closed_at, interval, working_days, date_created, date_updated
	FROM
		general_agenda
	WHERE
		business_id = :business_id
	`

	var dbgAgd dbGeneralAgenda
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbgAgd); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return agenda.GeneralAgenda{}, fmt.Errorf("namedquerystruct: %w", agenda.ErrNotFound)
		}
		return agenda.GeneralAgenda{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	agd, err := toCoreGeneralAgenda(dbgAgd)
	if err != nil {
		return agenda.GeneralAgenda{}, err
	}

	return agd, nil
}

func (s *Store) QueryGeneralAgendaByID(ctx context.Context, agdID uuid.UUID) (agenda.GeneralAgenda, error) {
	data := struct {
		AgendaID string `db:"agenda_id"`
	}{
		AgendaID: agdID.String(),
	}

	const q = `
	SELECT 	
		id, business_id, opens_at, closed_at, interval, working_days, date_created, date_updated
	FROM
		general_agenda
	WHERE
		id = :agenda_id
	`

	var dbgAgd dbGeneralAgenda
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbgAgd); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return agenda.GeneralAgenda{}, fmt.Errorf("namedquerystruct: %w", agenda.ErrNotFound)
		}
		return agenda.GeneralAgenda{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	agd, err := toCoreGeneralAgenda(dbgAgd)
	if err != nil {
		return agenda.GeneralAgenda{}, err
	}

	return agd, nil
}

func (s *Store) CountGeneralAgenda(ctx context.Context, filter agenda.GAQueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1)
	FROM
		general_agenda	
	`

	buf := bytes.NewBufferString(q)
	s.applyFilterGeneralAgenda(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := db.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

func (s *Store) CreateDailyAgenda(ctx context.Context, agd agenda.DailyAgenda) error {
	const q = `
	INSERT INTO daily_agenda
		(id, business_id, opens_at, closed_at, interval, availability, date_created, date_updated)
	VALUES
		(:id, :business_id, :opens_at, :closed_at, :interval, :availability, :date_created, :date_updated)
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBDailyAgenda(agd)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) UpdateDailyAgenda(ctx context.Context, agd agenda.DailyAgenda) error {
	const q = `
	UPDATE
		daily_agenda
	SET
		"opens_at" = :opens_at,
		"closed_at" = :closed_at,
		"interval" = :interval,
		"date_updated" = :date_updated
	WHERE
		"id" = :id
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBDailyAgenda(agd)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) DeleteDailyAgenda(ctx context.Context, agd agenda.DailyAgenda) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: agd.ID.String(),
	}

	const q = `
	DELETE FROM
		daily_agenda	
	WHERE
		"id" = :id
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexedcontext: %w", err)
	}

	return nil
}

func (s *Store) QueryDailyAgenda(ctx context.Context, filter agenda.DAQueryFilter, orderBy order.By, page page.Page) ([]agenda.DailyAgenda, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT 	
		id, business_id, opens_at, closed_at, interval, availability, date_created, date_updated
	FROM
		daily_agenda
	`

	buf := bytes.NewBufferString(q)
	s.applyFilterDailyAgenda(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbdAgd []dbDailyAgenda
	if err := db.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbdAgd); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	agds, err := toCoreDailyAgendaSlice(dbdAgd)
	if err != nil {
		return nil, err
	}

	return agds, nil

}

func (s *Store) CountDailyAgenda(ctx context.Context, filter agenda.DAQueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1)
	FROM
		daily_agenda	
	`

	buf := bytes.NewBufferString(q)
	s.applyFilterDailyAgenda(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := db.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryDailyAgendaByID(ctx context.Context, agdID uuid.UUID) (agenda.DailyAgenda, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: agdID.String(),
	}

	const q = `
	SELECT 	
		id, business_id, opens_at, closed_at, interval, availability, date_created, date_updated
	FROM 
		daily_agenda
	WHERE
		id = :id
	`

	var agd dbDailyAgenda
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &agd); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return agenda.DailyAgenda{}, fmt.Errorf("namedquerystruct: %w", agenda.ErrNotFound)
		}
		return agenda.DailyAgenda{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	cAgd, err := toCoreDailyAgenda(agd)
	if err != nil {
		return agenda.DailyAgenda{}, err
	}

	return cAgd, nil
}
