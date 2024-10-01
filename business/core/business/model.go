package business

import (
	"time"

	"github.com/google/uuid"
)

type Business struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Desc        string
	DateCreated time.Time
	DateUpdated time.Time
}

type NewBusiness struct {
	OwnerID uuid.UUID
	Name    string
	Desc    string
}

type UpdateBusiness struct {
	Name *string
	Desc *string
}
