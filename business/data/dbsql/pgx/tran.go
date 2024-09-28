package db

import (
	"fmt"

	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/jmoiron/sqlx"
)

type dbBegineer struct {
	sqlxDB *sqlx.DB
}

func NewBeginner(sqlxDB *sqlx.DB) transaction.Beginner {
	return &dbBegineer{
		sqlxDB: sqlxDB,
	}
}

func (db *dbBegineer) Begin() (transaction.Transaction, error) {
	return db.sqlxDB.Beginx()
}

func GetExtContext(tx transaction.Transaction) (sqlx.ExtContext, error) {
	ec, ok := tx.(sqlx.ExtContext)
	if !ok {
		return nil, fmt.Errorf("Transactor(%T) not of a type *sqlx.Tx", tx)
	}

	return ec, nil
}
