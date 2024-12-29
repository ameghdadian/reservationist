package userdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/ameghdadian/service/business/core/user"
	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/google/uuid"

	"github.com/ameghdadian/service/foundation/logger"
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

func (s *Store) ExecuteUnderTransaction(tx transaction.Transaction) (user.Storer, error) {
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

func (s *Store) Create(ctx context.Context, usr user.User) error {
	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, phone_no, enabled, date_created, date_updated)	
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :phone_no, :enabled, :date_created, :date_updated)
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, db.ErrDBDuplicateEntry) {
			return fmt.Errorf("namedexeccontext: %w", user.ErrUniqueEmailOrPhoneNo)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, usr user.User) error {
	const q = `
	UPDATE
		users
	SET
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"enabled" = :enabled,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id
	`
	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, db.ErrDBDuplicateEntry) {
			return user.ErrUniqueEmailOrPhoneNo
		}
		return fmt.Errorf("namedexedcontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, usr user.User) error {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: usr.ID.String(),
	}
	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id
	`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter user.QueryFilter, orderBy order.By, page page.Page) ([]user.User, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT	
		user_id, name, email, password_hash, roles, phone_no, enabled, date_created, date_updated	
	FROM
		users
	`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbUsrs []dbUser
	if err := db.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbUsrs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	usrs, err := toCoreUserSlice(dbUsrs)
	if err != nil {
		return nil, err
	}

	return usrs, nil
}

func (s *Store) Count(ctx context.Context, filter user.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT	
		count(1)
	FROM
		users
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

func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email.Address,
	}

	const q = `
		SELECT 	
			user_id, name, email, password_hash, roles, phone_no, enabled, date_created, date_updated	
		FROM
			users
		WHERE
			email = :email
	`

	var dbUsr dbUser
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return user.User{}, fmt.Errorf("namedquerystruct: %w", user.ErrNotFound)
		}
		return user.User{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	usr, err := toCoreUser(dbUsr)
	if err != nil {
		return user.User{}, err
	}

	return usr, err
}

func (s *Store) QueryByIDs(ctx context.Context, userID []uuid.UUID) ([]user.User, error) {
	data := struct {
		IDs []uuid.UUID `db:"user_ids"`
	}{
		IDs: userID,
	}

	const q = `
	SELECT	
		user_id, name, email, password_hash, roles, phone_no, enabled, date_created, date_updated	
	FROM
		users
	WHERE
		user_id = ANY (:user_ids)
	`

	var dbUsrs []dbUser
	if err := db.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbUsrs); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return nil, user.ErrNotFound
		}

		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	usrs, err := toCoreUserSlice(dbUsrs)
	if err != nil {
		return nil, err
	}

	return usrs, nil
}

func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT 
		user_id, name, email, password_hash, roles, phone_no, enabled, date_created, date_updated	
	FROM
		users
	WHERE
		user_id = :user_id
	`

	var dbUsr dbUser
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return user.User{}, fmt.Errorf("namedquerystruct: %w", user.ErrNotFound)
		}
		return user.User{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	usr, err := toCoreUser(dbUsr)
	if err != nil {
		return user.User{}, err
	}

	return usr, nil
}
