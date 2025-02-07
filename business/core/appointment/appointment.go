package appointment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/otel"
	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("appointment not found")
	ErrUserDisabled     = errors.New("user disabled")
	ErrPastTime         = errors.New("time past now")
	ErrAlreadyCancelled = errors.New("appointment already cancelled")
	ErrAlreadyReserved  = errors.New("given time is already reserverd")
)

type Storer interface {
	ExecuteUnderTransaction(tx transaction.Transaction) (Storer, error)
	Create(ctx context.Context, apt Appointment) error
	Update(ctx context.Context, apt Appointment) error
	Delete(ctx context.Context, apt Appointment) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Appointment, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, aptID uuid.UUID) (Appointment, error)
	QueryByUserID(ctx context.Context, usrID uuid.UUID) ([]Appointment, error)
	QueryByBusinessID(ctx context.Context, bsnID uuid.UUID) ([]Appointment, error)
}

type Core struct {
	storer  Storer
	log     *logger.Logger
	usrCore *user.Core
	bsnCore *business.Core
	task    *Task
}

func NewCore(log *logger.Logger, usrCore *user.Core, bsnCore *business.Core, storer Storer, task *Task) *Core {
	return &Core{
		storer:  storer,
		log:     log,
		usrCore: usrCore,
		bsnCore: bsnCore,
		task:    task,
	}
}

func (c *Core) ExecuteUnderTransaction(tx transaction.Transaction) (*Core, error) {
	storer, err := c.storer.ExecuteUnderTransaction(tx)
	if err != nil {
		return nil, err
	}

	usrCore, err := c.usrCore.ExecuteUnderTransaction(tx)
	if err != nil {
		return nil, err
	}

	bsnCore, err := c.bsnCore.ExecuteUnderTransaction(tx)
	if err != nil {
		return nil, err
	}

	c = &Core{
		storer:  storer,
		log:     c.log,
		usrCore: usrCore,
		bsnCore: bsnCore,
		task:    c.task,
	}

	return c, nil
}

func (c *Core) Create(ctx context.Context, na NewAppointment) (Appointment, error) {
	ctx, span := otel.AddSpan(ctx, "business.appointment.create")
	defer span.End()

	usr, err := c.usrCore.QueryByID(ctx, na.UserID)
	if err != nil {
		return Appointment{}, fmt.Errorf("user.querybyid: %s: %w", na.UserID, err)
	}

	if !usr.Enabled {
		return Appointment{}, ErrUserDisabled
	}

	bsn, err := c.bsnCore.QueryByID(ctx, na.BusinessID)
	if err != nil {
		return Appointment{}, fmt.Errorf("business.querybyid: %s: %w", na.BusinessID, err)
	}

	if na.ScheduledOn.UTC().Before(time.Now().UTC()) {
		return Appointment{}, ErrPastTime
	}

	var filter QueryFilter
	filter.WithScheduledOn(na.ScheduledOn.UTC())
	filter.WithBusinessID(na.BusinessID)

	page, err := page.Parse("1", "1")
	if err != nil {
		return Appointment{}, fmt.Errorf("couldn't parse page parameters: %w", err)
	}

	agds, err := c.storer.Query(ctx, filter, DefaultOrderBy, page)
	if err != nil {
		return Appointment{}, fmt.Errorf("query: %w", err)
	}
	if len(agds) == 1 {
		return Appointment{}, ErrAlreadyReserved
	}
	if len(agds) > 1 {
		return Appointment{}, fmt.Errorf("found unexpected number of entries(gt. 1)")
	}

	now := time.Now()

	apt := Appointment{
		ID:          uuid.New(),
		BusinessID:  bsn.ID,
		UserID:      usr.ID,
		Status:      na.Status,
		ScheduledOn: na.ScheduledOn,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := c.storer.Create(ctx, apt); err != nil {
		return Appointment{}, fmt.Errorf("create: %w", err)
	}

	_, err = c.task.NewSendSMSTask(usr.ID, na.ScheduledOn, apt.ID.String())
	if err != nil {
		return Appointment{}, fmt.Errorf("newsendsmstask: %w", err)
	}

	return apt, nil
}

func (c *Core) Update(ctx context.Context, apt Appointment, uapt UpdateAppointment) (Appointment, error) {
	// TODO: Check given scheduled time is not already booked for this business.
	// It's already done once in Create Method above. See if you can Copy/Paste it.

	ctx, span := otel.AddSpan(ctx, "business.appointment.update")
	defer span.End()

	// Query appointment to check if the scheduled time is not passed
	// or appointment is not already cancelled
	if apt.ScheduledOn.UTC().Before(time.Now().UTC()) {
		return Appointment{}, ErrPastTime
	}

	if apt.Status == StatusCancelled {
		return Appointment{}, ErrAlreadyCancelled
	}

	if uapt.Status != nil {
		apt.Status = *uapt.Status
	}

	if uapt.ScheduledOn != nil {
		apt.ScheduledOn = *uapt.ScheduledOn

		_, err := c.task.updateSendSMSTask(apt.UserID, *uapt.ScheduledOn, apt.ID.String())
		if err != nil {
			return Appointment{}, fmt.Errorf("updatesendsmstask: %w", err)
		}
	}

	apt.DateUpdated = time.Now()
	if err := c.storer.Update(ctx, apt); err != nil {
		return Appointment{}, fmt.Errorf("update: %w", err)
	}

	return apt, nil
}

func (c *Core) Delete(ctx context.Context, apt Appointment) error {
	ctx, span := otel.AddSpan(ctx, "business.appointment.delete")
	defer span.End()

	if err := c.storer.Delete(ctx, apt); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := c.task.cancelSendSMSTask(apt.ID.String()); err != nil {
		return fmt.Errorf("cancelsendsmstask: %w", err)
	}

	return nil
}

func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Appointment, error) {
	ctx, span := otel.AddSpan(ctx, "business.appointment.query")
	defer span.End()

	apts, err := c.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return apts, nil
}

func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.appointment.count")
	defer span.End()

	return c.storer.Count(ctx, filter)
}

func (c *Core) QueryByID(ctx context.Context, aptID uuid.UUID) (Appointment, error) {
	ctx, span := otel.AddSpan(ctx, "business.appointment.querybyid")
	defer span.End()

	apt, err := c.storer.QueryByID(ctx, aptID)
	if err != nil {
		return Appointment{}, fmt.Errorf("query: appointmentID[%s]: %w", aptID, err)
	}

	return apt, nil
}

func (c *Core) QueryByUserID(ctx context.Context, usrID uuid.UUID) ([]Appointment, error) {
	ctx, span := otel.AddSpan(ctx, "business.appointment.querybyuserid")
	defer span.End()

	apts, err := c.storer.QueryByUserID(ctx, usrID)
	if err != nil {
		return nil, fmt.Errorf("query: userID[%s]: %w", usrID, err)
	}

	return apts, nil
}

func (c *Core) QueryByBusinessID(ctx context.Context, bsnID uuid.UUID) ([]Appointment, error) {
	ctx, span := otel.AddSpan(ctx, "business.appointment.querybybusinessid")
	defer span.End()

	apts, err := c.storer.QueryByBusinessID(ctx, bsnID)
	if err != nil {
		return nil, fmt.Errorf("query: businessID[%s]: %w", bsnID, err)
	}

	return apts, nil
}
