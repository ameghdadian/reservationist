package businessgrp

import (
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/google/uuid"
)

type AppBusiness struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DateCreated string `json:"-"`
	DateUpdated string `json:"-"`
}

func toAppBusiness(b business.Business) AppBusiness {
	return AppBusiness{
		ID:          b.ID.String(),
		OwnerID:     b.OwnerID.String(),
		Name:        b.Name,
		Description: b.Desc,
		DateCreated: b.DateCreated.Format(time.RFC3339),
		DateUpdated: b.DateUpdated.Format(time.RFC3339),
	}
}

func toAppBusinesses(bsns []business.Business) []AppBusiness {
	items := make([]AppBusiness, len(bsns))
	for i, b := range bsns {
		items[i] = toAppBusiness(b)
	}

	return items
}

// ======================================================================

type AppNewBusiness struct {
	OwnerID     string `json:"owner_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required,max=140"`
}

func (app AppNewBusiness) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}

func toCoreNewBusiness(app AppNewBusiness) (business.NewBusiness, error) {
	ownerID, err := uuid.Parse(app.OwnerID)
	if err != nil {
		return business.NewBusiness{}, fmt.Errorf("parsing ownerID: %w", err)
	}

	nb := business.NewBusiness{
		OwnerID: ownerID,
		Name:    app.Name,
		Desc:    app.Description,
	}

	return nb, nil
}

// ======================================================================

type AppUpdateBusiness struct {
	Name *string `json:"name"`
	Desc *string `json:"description" validate:"omitempty,max=140"`
}

func (app AppUpdateBusiness) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}

func toCoreUpdateBusiness(app AppUpdateBusiness) business.UpdateBusiness {
	core := business.UpdateBusiness{
		Name: app.Name,
		Desc: app.Desc,
	}

	return core
}

// ======================================================================
