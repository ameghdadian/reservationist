package agenda

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/otel"
	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("agenda is not found")
	ErrOutOfRange     = errors.New("selected time is not within business working hours")
	ErrIntervalAbused = errors.New("interval is not respected")
	ErrNoDailyAgenda  = errors.New("no daily agenda found")
	ErrBusinessOff    = errors.New("business has no activity at given date")
)

type Storer interface {
	ExecuteUnderTransaction(tx transaction.Transaction) (Storer, error)
	CreateGeneralAgenda(ctx context.Context, agd GeneralAgenda) error
	UpdateGeneralAgenda(ctx context.Context, agd GeneralAgenda) error
	DeleteGeneralAgenda(ctx context.Context, agd GeneralAgenda) error
	QueryGeneralAgenda(ctx context.Context, filter GAQueryFilter, orderBy order.By, page page.Page) ([]GeneralAgenda, error)
	QueryGeneralAgendaByBusinessID(ctx context.Context, bsnID uuid.UUID) (GeneralAgenda, error)
	QueryGeneralAgendaByID(ctx context.Context, agdID uuid.UUID) (GeneralAgenda, error)
	CountGeneralAgenda(ctx context.Context, filter GAQueryFilter) (int, error)

	CreateDailyAgenda(ctx context.Context, agd DailyAgenda) error
	UpdateDailyAgenda(ctx context.Context, agd DailyAgenda) error
	DeleteDailyAgenda(ctx context.Context, agd DailyAgenda) error
	QueryDailyAgenda(ctx context.Context, filter DAQueryFilter, orderBy order.By, page page.Page) ([]DailyAgenda, error)
	CountDailyAgenda(ctx context.Context, filter DAQueryFilter) (int, error)
	QueryDailyAgendaByID(ctx context.Context, agdID uuid.UUID) (DailyAgenda, error)
}

type Core struct {
	storer  Storer
	bsnCore *business.Core
	log     *logger.Logger
}

func NewCore(log *logger.Logger, bsnCore *business.Core, storer Storer) *Core {
	return &Core{
		storer:  storer,
		bsnCore: bsnCore,
		log:     log,
	}
}

func (c *Core) ExecuteUnderTransaction(tx transaction.Transaction) (*Core, error) {
	storer, err := c.storer.ExecuteUnderTransaction(tx)
	if err != nil {
		return nil, err
	}

	c = &Core{
		storer: storer,
		log:    c.log,
	}

	return c, nil
}

func (c *Core) CreateGeneralAgenda(ctx context.Context, na NewGeneralAgenda) (GeneralAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.create")
	defer span.End()

	now := time.Now()

	agd := GeneralAgenda{
		ID:          uuid.New(),
		BusinessID:  na.BusinessID,
		OpensAt:     na.OpensAt,
		ClosedAt:    na.ClosedAt,
		Interval:    na.Interval,
		WorkingDays: na.WorkingDays,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := c.storer.CreateGeneralAgenda(ctx, agd); err != nil {
		return GeneralAgenda{}, fmt.Errorf("create general agenda: %w", err)
	}

	return agd, nil
}

func (c *Core) UpdateGenralAgenda(ctx context.Context, agd GeneralAgenda, uAgd UpdateGeneralAgenda) (GeneralAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.update")
	defer span.End()

	if uAgd.OpensAt != nil {
		agd.OpensAt = *uAgd.OpensAt
	}

	if uAgd.ClosedAt != nil {
		agd.ClosedAt = *uAgd.ClosedAt
	}

	if uAgd.Interval != nil {
		agd.Interval = *uAgd.Interval
	}

	if uAgd.WorkingDays != nil {
		agd.WorkingDays = uAgd.WorkingDays
	}

	agd.DateUpdated = time.Now()
	if err := c.storer.UpdateGeneralAgenda(ctx, agd); err != nil {
		return GeneralAgenda{}, fmt.Errorf("update: %w", err)
	}

	return agd, nil
}

func (c *Core) DeleteGeneralAgenda(ctx context.Context, agd GeneralAgenda) error {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.delete")
	defer span.End()

	if err := c.storer.DeleteGeneralAgenda(ctx, agd); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c *Core) QueryGeneralAgenda(ctx context.Context, filter GAQueryFilter, orderBy order.By, page page.Page) ([]GeneralAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.query")
	defer span.End()

	agds, err := c.storer.QueryGeneralAgenda(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return agds, nil
}

func (c *Core) QueryGeneralAgendaByBusinessID(ctx context.Context, bsnID uuid.UUID) (GeneralAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.querybybusinessid")
	defer span.End()

	agd, err := c.storer.QueryGeneralAgendaByBusinessID(ctx, bsnID)
	if err != nil {
		return GeneralAgenda{}, fmt.Errorf("query: bsnID[%s]: %w", bsnID, err)
	}

	return agd, nil
}

func (c *Core) QueryGeneralAgendaByID(ctx context.Context, agdID uuid.UUID) (GeneralAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.querybyid")
	defer span.End()

	agd, err := c.storer.QueryGeneralAgendaByID(ctx, agdID)
	if err != nil {
		return GeneralAgenda{}, fmt.Errorf("query: agdID[%s]: %w", agdID, err)
	}

	return agd, nil
}

func (c *Core) CountGeneralAgenda(ctx context.Context, filter GAQueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.count")
	defer span.End()

	return c.storer.CountGeneralAgenda(ctx, filter)
}

// conformsGeneralAgendaBoundary checks two things:
// 1. Whether check time is placed inside inclusive opening and closing business agenda,
// 2. And if check time conforms with interval requirement.
func (c *Core) conformGeneralAgendaBoundary(ctx context.Context, bsnID uuid.UUID, checkTime time.Time) error {
	ctx, span := otel.AddSpan(ctx, "business.generalagenda.conformboundary")
	defer span.End()

	agd, err := c.storer.QueryGeneralAgendaByBusinessID(ctx, bsnID)
	if err != nil {
		return fmt.Errorf("query: bsnID[%s]: %w", bsnID, err)
	}

	opens := agd.OpensAt.UTC()
	closed := agd.ClosedAt.UTC()
	check := checkTime.UTC()

	if check.Before(opens) || check.Equal(closed) || check.After(closed) {
		return ErrOutOfRange
	}

	interval := agd.Interval
	checkpoint := check.Sub(opens).Seconds()

	if int(math.Ceil(checkpoint))%interval != 0 {
		return ErrIntervalAbused
	}

	return nil
}

// -------------------------------------------------------------------------------------------------------

func (c *Core) CreateDailyAgenda(ctx context.Context, na NewDailyAgenda) (DailyAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.dailyagenda.create")
	defer span.End()

	now := time.Now()

	agd := DailyAgenda{
		ID:           uuid.New(),
		BusinessID:   na.BusinessID,
		OpensAt:      na.OpensAt,
		ClosedAt:     na.ClosedAt,
		Interval:     na.Interval,
		Availability: na.Availability,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := c.storer.CreateDailyAgenda(ctx, agd); err != nil {
		return DailyAgenda{}, fmt.Errorf("create daily agenda: %w", err)
	}

	return agd, nil
}

func (c *Core) UpdateDailyAgenda(ctx context.Context, agd DailyAgenda, uAgd UpdateDailyAgenda) (DailyAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.dailyagenda.update")
	defer span.End()

	if uAgd.OpensAt != nil {
		agd.OpensAt = *uAgd.OpensAt
	}

	if uAgd.ClosedAt != nil {
		agd.ClosedAt = *uAgd.ClosedAt
	}

	if uAgd.Interval != nil {
		agd.Interval = *uAgd.Interval
	}

	if uAgd.Availability != nil {
		agd.Availability = *uAgd.Availability
	}

	agd.DateUpdated = time.Now()

	if err := c.storer.UpdateDailyAgenda(ctx, agd); err != nil {
		return DailyAgenda{}, fmt.Errorf("update: %w", err)
	}

	return agd, nil
}

func (c *Core) DeleteDailyAgenda(ctx context.Context, agd DailyAgenda) error {
	ctx, span := otel.AddSpan(ctx, "business.dailyagenda.delete")
	defer span.End()

	if err := c.storer.DeleteDailyAgenda(ctx, agd); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c *Core) QueryDailyAgenda(ctx context.Context, filter DAQueryFilter, orderBy order.By, page page.Page) ([]DailyAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.dailyagenda.query")
	defer span.End()

	agds, err := c.storer.QueryDailyAgenda(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return agds, nil
}

func (c *Core) CountDailyAgenda(ctx context.Context, filter DAQueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.dailyagenda.count")
	defer span.End()

	return c.storer.CountDailyAgenda(ctx, filter)
}

func (c *Core) QueryDailyAgendaByID(ctx context.Context, agdID uuid.UUID) (DailyAgenda, error) {
	ctx, span := otel.AddSpan(ctx, "business.dailyagenda.querybyid")
	defer span.End()

	agd, err := c.storer.QueryDailyAgendaByID(ctx, agdID)
	if err != nil {
		return DailyAgenda{}, fmt.Errorf("query: dailyAgendaID[%s]: %w", agdID, err)
	}

	return agd, nil
}

// conformsDailyAgendaBoundary checks two things:
// 1. Whether check time is placed inside inclusive opening and closing business agenda,
// 2. And if check time conforms with interval requirement.
func (c *Core) conformDailyAgendaBoundary(ctx context.Context, bsnID uuid.UUID, checkTime time.Time) error {
	ctx, span := otel.AddSpan(ctx, "business.dailyagenda.conformboundary")
	defer span.End()

	var filter DAQueryFilter
	filter.WithBusinessID(bsnID)
	filter.WithDate(checkTime.UTC())

	pagination, err := page.Parse("1", "10")
	if err != nil {
		return fmt.Errorf("couldn't parse page parameters: %w", err)
	}
	agds, err := c.storer.QueryDailyAgenda(ctx, filter, DefaultOrderBy, pagination)
	if err != nil {
		return err
	}

	if len(agds) == 0 {
		return ErrNoDailyAgenda
	}

	err = ErrBusinessOff
	for i := range agds {
		if agds[i].Availability {

			opens := agds[i].OpensAt.UTC()
			closed := agds[i].ClosedAt.UTC()
			check := checkTime.UTC()

			if (check.Equal(opens) || check.After(opens)) && check.Before(closed) {

				interval := agds[i].Interval
				checkpoint := check.Sub(opens).Seconds()

				if int(math.Ceil(checkpoint))%interval != 0 {
					err = ErrIntervalAbused
					break
				}

				return nil
			}

			err = ErrOutOfRange
		}

	}

	return err
}

// -------------------------------------------------------------------------------------------------------

func (c *Core) TimeWithinAgendaBoundary(ctx context.Context, bsnID uuid.UUID, checkTime time.Time) error {
	err := c.conformDailyAgendaBoundary(ctx, bsnID, checkTime)
	if err != nil {
		if !errors.Is(err, ErrNoDailyAgenda) {
			return errs.New(errs.InvalidArgument, err)
		}

		// If doesn't conform with daily agenda, check with the general agenda to see any match.
		if err = c.conformGeneralAgendaBoundary(ctx, bsnID, checkTime); err != nil {
			return errs.New(errs.InvalidArgument, err)
		}
	}

	return nil
}
