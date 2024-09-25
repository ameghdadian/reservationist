package userdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/ameghdadian/service/business/core/user"
	db "github.com/ameghdadian/service/business/data/dbsql/pgx"

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
