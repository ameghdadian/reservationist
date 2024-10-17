package agenda

import (
	"time"

	"github.com/google/uuid"
)

// GeneralAgenda is the general detailed availability of a business during a week
// REMINDER: All fields are mandatory
type GeneralAgenda struct {
	ID         uuid.UUID
	BusinessID uuid.UUID
	// Parse this with RFC3339, then split it into time and timezone and store only these two inside psql 'time' type
	OpensAt     time.Time // At App layer, receive seconds from the start of the day, then add this to default time.Time value and store it in UTC timezone
	ClosedAt    time.Time // At App layer, receive secondds from the start of the day, then add this to default time.Time value and store it in UTC timezone
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
	ID         uuid.UUID
	BusinessID uuid.UUID
	OpensAt    time.Time // OPTIONAL
	ClosedAt   time.Time // OPTIONAL
	Interval   int       // OPTIONAL
	// MANDATORY; FORMAT IT USING time.DateOnly and store it this way in date type of psql,
	// Otherwise use https://github.com/googleapis/google-cloud-go/blob/v0.115.1/civil/civil.go#L253
	Date         time.Time
	Availability bool // MANDATORY
	DateCreated  time.Time
	DateUpdated  time.Time
}

type NewDailyAgenda struct {
	BusinessID   uuid.UUID
	OpensAt      time.Time
	ClosedAt     time.Time
	Interval     int
	Date         time.Time
	Availability bool
}

type UpdateDailyAgenda struct {
	OpensAt      *time.Time
	ClosedAt     *time.Time
	Interval     *int
	Date         *time.Time
	Availability *bool
}
