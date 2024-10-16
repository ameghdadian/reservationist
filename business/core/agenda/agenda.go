package agenda

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("agenda is not found")
)

type Storer interface {
	ExecuteUnderTransaction(tx transaction.Transaction) (Storer, error)
	CreateGeneralAgenda(ctx context.Context, agd GeneralAgenda) error
	UpdateGeneralAgenda(ctx context.Context, agd GeneralAgenda) error
	DeleteGeneralAgenda(ctx context.Context, agd GeneralAgenda) error
	QueryGeneralAgenda(ctx context.Context, filter GAQueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]GeneralAgenda, error)
	QueryGeneralAgendaByID(ctx context.Context, agdID uuid.UUID) (GeneralAgenda, error)
	CountGeneralAgenda(ctx context.Context, filter GAQueryFilter) (int, error)

	CreateDailyAgenda(ctx context.Context, agd DailyAgenda) error
	UpdateDailyAgenda(ctx context.Context, agd DailyAgenda) error
	DeleteDailyAgenda(ctx context.Context, agd DailyAgenda) error
	QueryDailyAgenda(ctx context.Context, filter DAQueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]DailyAgenda, error)
	CountDailyAgenda(ctx context.Context, filter DAQueryFilter) (int, error)
	QueryDailyAgendaByID(ctx context.Context, agdID uuid.UUID) (DailyAgenda, error)
}

type Core struct {
	storer  Storer
	log     *logger.Logger
	bsnCore *business.Core
}

func NewCore(log *logger.Logger, bsnCore *business.Core, storer Storer) *Core {
	return &Core{
		storer:  storer,
		log:     log,
		bsnCore: bsnCore,
	}
}

func (c *Core) ExecuteUnderTransaction(tx transaction.Transaction) (*Core, error) {

	return nil, nil
}

func (c *Core) CreateGeneralAgenda(ctx context.Context, na NewGeneralAgenda) (GeneralAgenda, error) {
	// TODO: Create an AuthorizeGeneralAgenda middleware to check only users who own a business can create agenda plan for it.

	_, err := c.bsnCore.QueryByID(ctx, na.BusinessID)
	if err != nil {
		// TODO: DOUBLE Check this!! We don't want this end up being 500 error!
		return GeneralAgenda{}, fmt.Errorf("busineess.querybyid: %s: %w", na.BusinessID, err)
	}

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
	if err := c.storer.DeleteGeneralAgenda(ctx, agd); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c *Core) QueryGeneralAgenda(ctx context.Context, filter GAQueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]GeneralAgenda, error) {
	agds, err := c.storer.QueryGeneralAgenda(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return agds, nil
}

func (c *Core) QueryGeneralAgendaByID(ctx context.Context, agdID uuid.UUID) (GeneralAgenda, error) {
	agd, err := c.storer.QueryGeneralAgendaByID(ctx, agdID)
	if err != nil {
		return GeneralAgenda{}, fmt.Errorf("query: agdID[%s]: %w", agdID, err)
	}

	return agd, nil
}

func (c *Core) CountGeneralAgenda(ctx context.Context, filter GAQueryFilter) (int, error) {
	return c.storer.CountGeneralAgenda(ctx, filter)
}

// -------------------------------------------------------------------------------------------------------

func (c *Core) CreateDailyAgenda(ctx context.Context, na NewDailyAgenda) (DailyAgenda, error) {
	// TODO: Create an AuthorizeDailyAgenda middleware to check only users who own a business can create agenda plan for it.

	_, err := c.bsnCore.QueryByID(ctx, na.BusinessID)
	if err != nil {
		// TODO: DOUBLE Check this!! We don't want this end up being 500 error!
		return DailyAgenda{}, fmt.Errorf("busineess.querybyid: %s: %w", na.BusinessID, err)
	}

	now := time.Now()

	agd := DailyAgenda{
		ID:           uuid.New(),
		BusinessID:   na.BusinessID,
		OpensAt:      na.OpensAt,
		ClosedAt:     na.ClosedAt,
		Interval:     na.Interval,
		Date:         na.Date,
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
	if uAgd.OpensAt != nil {
		agd.OpensAt = *uAgd.OpensAt
	}

	if uAgd.ClosedAt != nil {
		agd.ClosedAt = *uAgd.ClosedAt
	}

	if uAgd.Interval != nil {
		agd.Interval = *uAgd.Interval
	}

	if uAgd.Date != nil {
		agd.Date = *uAgd.Date
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
	if err := c.storer.DeleteDailyAgenda(ctx, agd); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c *Core) QueryDailyAgenda(ctx context.Context, filter DAQueryFilter, orderBy order.By, page int, rowsPerPage int) ([]DailyAgenda, error) {
	agds, err := c.storer.QueryDailyAgenda(ctx, filter, orderBy, page, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return agds, nil
}

func (c *Core) CountDailyAgenda(ctx context.Context, filter DAQueryFilter) (int, error) {
	return c.storer.CountDailyAgenda(ctx, filter)
}

func (c *Core) QueryDailyAgendaByID(ctx context.Context, agdID uuid.UUID) (DailyAgenda, error) {
	agd, err := c.storer.QueryDailyAgendaByID(ctx, agdID)
	if err != nil {
		return DailyAgenda{}, fmt.Errorf("query: dailyAgendaID[%s]: %w", agdID, err)
	}

	return agd, nil
}
