package businessdb

import (
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/google/uuid"
)

type dbBusiness struct {
	ID          uuid.UUID `db:"business_id"`
	OwnerID     uuid.UUID `db:"owner_id"`
	Name        string    `db:"name"`
	Desc        string    `db:"description"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toDBBusiness(b business.Business) dbBusiness {
	return dbBusiness{
		ID:          b.ID,
		OwnerID:     b.OwnerID,
		Name:        b.Name,
		Desc:        b.Desc,
		DateCreated: b.DateCreated.UTC(),
		DateUpdated: b.DateUpdated.UTC(),
	}
}

func toCoreBusiness(dbBsn dbBusiness) business.Business {
	b := business.Business{
		ID:          dbBsn.ID,
		OwnerID:     dbBsn.OwnerID,
		Name:        dbBsn.Name,
		Desc:        dbBsn.Desc,
		DateCreated: dbBsn.DateCreated.In(time.Local),
		DateUpdated: dbBsn.DateUpdated.In(time.Local),
	}

	return b
}

func toCoreBusinessSlice(dbBsns []dbBusiness) []business.Business {
	bsns := make([]business.Business, len(dbBsns))
	for i, b := range dbBsns {
		bsns[i] = toCoreBusiness(b)
	}

	return bsns
}
