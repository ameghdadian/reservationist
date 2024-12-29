package agenda

import (
	"fmt"

	"github.com/ameghdadian/service/foundation/errs"
	"github.com/google/uuid"
)

type GAQueryFilter struct {
	ID          *uuid.UUID `validate:"omitempty,uuid"`
	BusinesesID *uuid.UUID `validate:"omitempty,uuid"`
}

func (qf *GAQueryFilter) Validate() error {
	if err := errs.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func (qf *GAQueryFilter) WithGenealAgendaID(id uuid.UUID) {
	qf.ID = &id
}

func (qf *GAQueryFilter) WithBusinessID(bsnID uuid.UUID) {
	qf.BusinesesID = &bsnID
}

// --------------------------------------------------------------------

type DAQueryFilter struct {
	ID         *uuid.UUID `validate:"omitempty,uuid"`
	BusinessID *uuid.UUID `validate:"omitempty,uuid"`
	Date       *string    `validadte:"omitempty,excluded_with=From To Days"`
	From       *string    `validate:"omitempty,required_with=To"`
	To         *string    `validate:"omitempty,required_with=From"`
	Days       *int       `validate:"omitempty,number,lte=30,excluded_with=From To Date"`
}

func (qf *DAQueryFilter) Validate() error {
	if err := errs.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func (qf *DAQueryFilter) WithDailyAgendaID(id uuid.UUID) {
	qf.ID = &id
}

func (qf *DAQueryFilter) WithBusinessID(id uuid.UUID) {
	qf.BusinessID = &id
}

func (qf *DAQueryFilter) WithDate(date string) {
	qf.Date = &date
}

func (qf *DAQueryFilter) WithFrom(from string) {
	qf.From = &from
}

func (qf *DAQueryFilter) WithTo(to string) {
	qf.To = &to
}

func (qf *DAQueryFilter) WithDays(days int) {
	qf.Days = &days
}
