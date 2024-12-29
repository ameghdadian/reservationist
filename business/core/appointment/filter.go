package appointment

import (
	"fmt"
	"time"

	"github.com/ameghdadian/service/foundation/errs"
	"github.com/google/uuid"
)

type QueryFilter struct {
	ID               *uuid.UUID `validate:"omitempty"`
	BusinessID       *uuid.UUID `validate:"omitempty"`
	UserID           *uuid.UUID `validate:"omitempty"`
	Status           *Status    `validate:"omitempty"`
	ScheduledOn      *time.Time `validate:"omitempty"`
	StartCreatedDate *time.Time `validate:"omitempty"`
	EndCreatedDate   *time.Time `validate:"omitempty"`
}

func (qf *QueryFilter) Validate() error {
	if err := errs.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func (qf *QueryFilter) WithAppointmentID(aptID uuid.UUID) {
	qf.ID = &aptID
}

func (qf *QueryFilter) WithBusinessID(bsnID uuid.UUID) {
	qf.BusinessID = &bsnID
}

func (qf *QueryFilter) WithUserID(usrID uuid.UUID) {
	qf.UserID = &usrID
}

func (qf *QueryFilter) WithStatus(status Status) {
	qf.Status = &status
}

func (qf *QueryFilter) WithScheduledOn(on time.Time) {
	d := on.UTC()
	qf.ScheduledOn = &d
}

func (qf *QueryFilter) WithStartCreatedDate(startDate time.Time) {
	d := startDate.UTC()
	qf.StartCreatedDate = &d
}

func (qf *QueryFilter) WithEndCreatedDate(endDate time.Time) {
	d := endDate.UTC()
	qf.EndCreatedDate = &d
}
