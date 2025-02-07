package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/ameghdadian/service/business/data/order"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/transaction"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/otel"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmailOrPhoneNo  = errors.New("email or phone number is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

type Storer interface {
	ExecuteUnderTransaction(tx transaction.Transaction) (Storer, error)
	Create(ctx context.Context, usr User) error
	Update(ctx context.Context, usr User) error
	Delete(ctx context.Context, usr User) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]User, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (User, error)
	QueryByIDs(ctx context.Context, userID []uuid.UUID) ([]User, error)
	QueryByEmail(ctx context.Context, email mail.Address) (User, error)
}

type Core struct {
	storer Storer
	log    *logger.Logger
}

func NewCore(log *logger.Logger, storer Storer) *Core {
	return &Core{
		storer: storer,
		log:    log,
	}
}

func (c *Core) ExecuteUnderTransaction(tx transaction.Transaction) (*Core, error) {
	trS, err := c.storer.ExecuteUnderTransaction(tx)
	if err != nil {
		return nil, err
	}

	c = &Core{
		storer: trS,
		log:    c.log,
	}

	return c, nil
}

func (c *Core) Create(ctx context.Context, nu NewUser) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.create")
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generatefrompassword: %w", err)
	}

	now := time.Now()

	usr := User{
		ID:           uuid.New(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		Enabled:      false,
		PhoneNo:      nu.PhoneNo,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := c.storer.Create(ctx, usr); err != nil {
		return User{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}

func (c *Core) Update(ctx context.Context, usr User, uu UpdateUser) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.update")
	defer span.End()

	if uu.Name != nil {
		usr.Name = *uu.Name
	}

	if uu.Email != nil {
		usr.Email = *uu.Email
	}

	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}

	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, fmt.Errorf("generatefrompassword: %w", err)
		}

		usr.PasswordHash = pw
	}

	if uu.Enabled != nil {
		usr.Enabled = *uu.Enabled
	}
	usr.DateUpdated = time.Now()

	if err := c.storer.Update(ctx, usr); err != nil {
		return User{}, fmt.Errorf("update: %w", err)
	}

	return usr, nil
}

func (c *Core) Delete(ctx context.Context, usr User) error {
	ctx, span := otel.AddSpan(ctx, "business.user.delete")
	defer span.End()

	if err := c.storer.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]User, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.query")
	defer span.End()

	users, err := c.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.count")
	defer span.End()

	return c.storer.Count(ctx, filter)
}

func (c *Core) QueryByID(ctx context.Context, userID uuid.UUID) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.querybyid")
	defer span.End()

	user, err := c.storer.QueryByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("query: userID[%s]: %w", userID, err)
	}

	return user, nil
}

func (c *Core) QueryByIDs(ctx context.Context, userID []uuid.UUID) ([]User, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.querybyids")
	defer span.End()

	usrs, err := c.storer.QueryByIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querybyids: ids[%q]: %w", userID, err)
	}

	return usrs, nil
}

func (c *Core) QueryByEmail(ctx context.Context, email mail.Address) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.querybyemail")
	defer span.End()

	user, err := c.storer.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s']: %w", email, err)
	}

	return user, nil
}

func (c *Core) Authenticate(ctx context.Context, email mail.Address, pass string) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.user.authenticate")
	defer span.End()

	usr, err := c.storer.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(pass)); err != nil {
		return User{}, fmt.Errorf("comparehashandpassword: %w", ErrAuthenticationFailure)
	}

	return usr, nil
}
