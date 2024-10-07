package appointment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func TestGenerateNewAppointment(n int, usrID uuid.UUID, bsnID uuid.UUID) []NewAppointment {
	na := make([]NewAppointment, n)
	for i := 0; i < n; i++ {
		na[i] = NewAppointment{
			BusinessID:  bsnID,
			UserID:      usrID,
			Status:      StatusScheduled,
			ScheduledOn: time.Now().Add(2 * time.Hour),
		}
	}

	return na
}

func TestGenerateSeedAppointments(n int, api *Core, usrID uuid.UUID, bsnID uuid.UUID) ([]Appointment, error) {
	newApts := TestGenerateNewAppointment(n, usrID, bsnID)

	apts := make([]Appointment, n)
	for i, na := range newApts {
		apt, err := api.Create(context.Background(), na)
		if err != nil {
			return nil, fmt.Errorf("seeding appointment: idx: %d: %w", i, err)
		}

		apts[i] = apt
	}

	return apts, nil
}
