package businessdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ameghdadian/service/business/core/business"
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

func (s *Store) ExecuteUnderTransaction(tx transaction.Transaction) (business.Storer, error) {
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

func (s *Store) Create(ctx context.Context, b business.Business) error {
	const q = `
	INSERT INTO businesses
		(business_id, owner_id, name, description, date_created, date_updated)
	VALUES
		(:business_id, :owner_id, :name, :description, :date_created, :date_updated)	
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBBusiness(b)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, b business.Business) error {
	const q = `
	UPDATE 
		businesses
	SET
		"name" = :name,
		"description" = :description
	WHERE
		business_id = :business_id
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBBusiness(b)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, b business.Business) error {
	data := struct {
		BusinessID string `db:"business_id"`
	}{
		BusinessID: b.ID.String(),
	}
	const q = `
	DELETE FROM
		businesses
	WHERE
		business_id = :business_id	
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter business.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]business.Business, error) {
	data := map[string]any{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `
	SELECT
		business_id, owner_id, name, description, date_created, date_updated
	FROM
		businesses
	`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbBsns []dbBusiness
	if err := db.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbBsns); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	bsns := toCoreBusinessSlice(dbBsns)

	return bsns, nil
}

func (s *Store) Count(ctx context.Context, filter business.QueryFilter) (int, error) {
	var data map[string]any

	const q = `
	SELECT 
		COUNT(1)
	FROM
		businesses
	`
	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, bsnID uuid.UUID) (business.Business, error) {
	data := struct {
		BusinessID string `db:"business_id"`
	}{
		BusinessID: bsnID.String(),
	}

	const q = `
	SELECT
		business_id, owner_id, name, description, date_created, date_updated
	FROM
		businesses
	WHERE
		business_id = :business_id
	`

	var dbBsn dbBusiness
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbBsn); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return business.Business{}, fmt.Errorf("namedquerystruct: %w", business.ErrNotFound)
		}
		return business.Business{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreBusiness(dbBsn), nil
}

func (s *Store) QueryByOwnerID(ctx context.Context, owrID uuid.UUID) ([]business.Business, error) {
	data := struct {
		OwnerID string `db:"owner_id"`
	}{
		OwnerID: owrID.String(),
	}

	const q = `
	SELECT
		business_id, owner_id, name, description, date_created, date_updated
	FROM
		businesses
	WHERE
		owner_id = :owner_id
	`

	var dbBsns []dbBusiness
	if err := db.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbBsns); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreBusinessSlice(dbBsns), nil
}
