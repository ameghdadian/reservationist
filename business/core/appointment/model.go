package appointment

import (
	"time"

	"github.com/google/uuid"
)

type Appointment struct {
	ID          uuid.UUID
	BusinessID  uuid.UUID
	UserID      uuid.UUID
	Status      Status
	ScheduledOn time.Time
	DateCreated time.Time
	DateUpdated time.Time
}

type NewAppointment struct {
	BusinessID  uuid.UUID
	UserID      uuid.UUID
	Status      Status
	ScheduledOn time.Time
}

type UpdateAppointment struct {
	Status      *Status
	ScheduledOn *time.Time
}
