package appointmentdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ameghdadian/service/business/core/appointment"
	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ameghdadian/service/business/data/order"
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

func (s *Store) ExecuteUnderTransaction(tx transaction.Transaction) (appointment.Storer, error) {
	ec, err := db.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	s = &Store{
		db:  ec,
		log: s.log,
	}

	return s, nil
}

func (s *Store) Create(ctx context.Context, apt appointment.Appointment) error {
	const q = `
	INSERT INTO appointments
		(appointment_id, business_id, user_id, status, scheduled_on, date_created, date_updated)
	VALUES
		(:appointment_id, :business_id, :user_id, :status, :scheduled_on, :date_created, :date_updated)
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBAppointment(apt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, apt appointment.Appointment) error {
	const q = `
	UPDATE
		appointments
	SET
		"status" = :status,
		"scheduled_on" = :scheduled_on,
		"date_updated" = :date_updated
	WHERE
		appointment_id = :appointment_id
	`
	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBAppointment(apt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, apt appointment.Appointment) error {
	data := struct {
		AppointmentID string `db:"appointment_id"`
	}{
		AppointmentID: apt.ID.String(),
	}

	const q = `
	DELETE FROM
		appointments
	WHERE
		appointment_id = :appointment_id	
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter appointment.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]appointment.Appointment, error) {
	data := map[string]any{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `
	SELECT	
		(appointment_id, business_id, user_id, status, scheduled_on, date_created, date_updated)
	FROM
		appointments
	`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ONLY")

	var dbApts []dbAppointment
	if err := db.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbApts); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	apts := toCoreAppointmentSlice(dbApts)

	return apts, nil
}

func (s *Store) Count(ctx context.Context, filter appointment.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1)
	FROM
		appointments	
	`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	if err := db.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, aptID uuid.UUID) (appointment.Appointment, error) {
	data := struct {
		AppointmentID string `db:"appointment_id"`
	}{
		AppointmentID: aptID.String(),
	}
	const q = `
	SELECT	
		appointment_id, business_id, user_id, status, scheduled_on, date_created, date_updated
	FROM
		appointments
	WHERE
		appointment_id = :appointment_id
	`

	var dbApt dbAppointment
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbApt); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return appointment.Appointment{}, fmt.Errorf("namedquerystruct: %w", appointment.ErrNotFound)
		}
		return appointment.Appointment{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreAppointment(dbApt), nil
}

func (s *Store) QueryByUserID(ctx context.Context, usrID uuid.UUID) ([]appointment.Appointment, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: usrID.String(),
	}

	const q = `
	SELECT
		appointment_id, business_id, user_id, status, scheduled_on, date_created, date_updated
	FROM
		appointments
	WHERE
		user_id = :user_id
	`

	var dbApts []dbAppointment
	if err := db.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbApts); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAppointmentSlice(dbApts), nil
}

func (s *Store) QueryByBusinessID(ctx context.Context, bsnID uuid.UUID) ([]appointment.Appointment, error) {
	data := struct {
		BusinessID string `db:"business_id"`
	}{
		BusinessID: bsnID.String(),
	}

	const q = `
	SELECT
		appointment_id, business_id, user_id, status, scheduled_on, date_created, date_updated
	FROM
		appointments
	WHERE
		business_id = :business_id
	`

	var dbApts []dbAppointment
	if err := db.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbApts); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreAppointmentSlice(dbApts), nil
}
