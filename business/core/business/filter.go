package business

import (
	"fmt"
	"time"

	"github.com/ameghdadian/service/foundation/errs"
	"github.com/google/uuid"
)

type QueryFilter struct {
	ID               *uuid.UUID `validate:"omitempty"`
	Name             *string    `validate:"omitempty"`
	Desc             *string    `validate:"omitempty"`
	StartCreatedDate *time.Time `validate:"omitempty"`
	EndCreatedDate   *time.Time `validate:"omitempty"`
}

func (qf *QueryFilter) Validate() error {
	if err := errs.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func (qf *QueryFilter) WithBusinessID(bsnID uuid.UUID) {
	qf.ID = &bsnID
}

func (qf *QueryFilter) WithName(name string) {
	qf.Name = &name
}
func (qf *QueryFilter) WithDesc(desc string) {
	qf.Desc = &desc
}
func (qf *QueryFilter) WithStartCreatedDate(startDate time.Time) {
	d := startDate.UTC()
	qf.StartCreatedDate = &d
}
func (qf *QueryFilter) WithEndCreatedDate(endDate time.Time) {
	d := endDate.UTC()
	qf.EndCreatedDate = &d
}
