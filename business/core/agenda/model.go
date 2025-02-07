package agenda

import (
	"time"

	"github.com/google/uuid"
)

// GeneralAgenda is the general detailed availability of a business during a week
// REMINDER: All fields are mandatory
type GeneralAgenda struct {
	ID          uuid.UUID
	BusinessID  uuid.UUID
	OpensAt     time.Time
	ClosedAt    time.Time
	Interval    int
	WorkingDays []Day
	DateCreated time.Time
	DateUpdated time.Time
}

type NewGeneralAgenda struct {
	BusinessID  uuid.UUID
	OpensAt     time.Time
	ClosedAt    time.Time
	Interval    int
	WorkingDays []Day
}

type UpdateGeneralAgenda struct {
	OpensAt     *time.Time
	ClosedAt    *time.Time
	Interval    *int
	WorkingDays []Day
}

// ------------------------------------------------------

// DailyAgenda represents daily modifications to general agenda which each business might need
type DailyAgenda struct {
	ID           uuid.UUID
	BusinessID   uuid.UUID
	OpensAt      time.Time // OPTIONAL
	ClosedAt     time.Time // OPTIONAL
	Interval     int       // OPTIONAL
	Availability bool      // MANDATORY
	DateCreated  time.Time
	DateUpdated  time.Time
}

type NewDailyAgenda struct {
	BusinessID   uuid.UUID
	OpensAt      time.Time
	ClosedAt     time.Time
	Interval     int
	Availability bool
}

type UpdateDailyAgenda struct {
	OpensAt      *time.Time
	ClosedAt     *time.Time
	Interval     *int
	Availability *bool
}
