package business

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("product not found")
	ErrUserDisabled = errors.New("user disabled")
)

type Storer interface {
	ExecuteUnderTransaction(tx transaction.Transaction) (Storer, error)
	Create(ctx context.Context, b Business) error
	Update(ctx context.Context, b Business) error
	Delete(ctx context.Context, b Business) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Business, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, bsnID uuid.UUID) (Business, error)
	QueryByOwnerID(ctx context.Context, owrID uuid.UUID) ([]Business, error)
}

type Core struct {
	storer  Storer
	log     *logger.Logger
	usrCore *user.Core
}

func NewCore(log *logger.Logger, usrCore *user.Core, storer Storer) *Core {
	return &Core{
		storer:  storer,
		usrCore: usrCore,
		log:     log,
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

	c = &Core{
		storer:  storer,
		log:     c.log,
		usrCore: usrCore,
	}

	return c, nil
}

func (c *Core) Create(ctx context.Context, nb NewBusiness) (Business, error) {
	usr, err := c.usrCore.QueryByID(ctx, nb.OwnerID)
	if err != nil {
		return Business{}, fmt.Errorf("user.querybyid: %s: %w", nb.OwnerID, err)
	}

	if !usr.Enabled {
		return Business{}, ErrUserDisabled
	}

	now := time.Now()

	bsn := Business{
		ID:          uuid.New(),
		OwnerID:     nb.OwnerID,
		Name:        nb.Name,
		Desc:        nb.Desc,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := c.storer.Create(ctx, bsn); err != nil {
		return Business{}, fmt.Errorf("create: %w", err)
	}

	return bsn, nil
}

func (c *Core) Update(ctx context.Context, b Business, ub UpdateBusiness) (Business, error) {
	if ub.Name != nil {
		b.Name = *ub.Name
	}

	if ub.Desc != nil {
		b.Desc = *ub.Desc
	}

	b.DateUpdated = time.Now()

	if err := c.storer.Update(ctx, b); err != nil {
		return Business{}, fmt.Errorf("update: %w", err)
	}

	return b, nil
}

func (c *Core) Delete(ctx context.Context, b Business) error {
	if err := c.storer.Delete(ctx, b); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Business, error) {
	bsns, err := c.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return bsns, nil
}

func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return c.storer.Count(ctx, filter)
}

func (c *Core) QueryByID(ctx context.Context, bsnID uuid.UUID) (Business, error) {
	bsn, err := c.storer.QueryByID(ctx, bsnID)
	if err != nil {
		return Business{}, fmt.Errorf("query: businessID[%s]: %w", bsnID, err)
	}

	return bsn, nil
}

func (c *Core) QueryByOwnerID(ctx context.Context, owrID uuid.UUID) ([]Business, error) {
	bsns, err := c.storer.QueryByOwnerID(ctx, owrID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return bsns, nil
}
