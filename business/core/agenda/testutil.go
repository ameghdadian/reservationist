package agenda

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

func TestGenerateNewGeneralAgendas(n int, bsnID uuid.UUID, userID uuid.UUID) ([]NewGeneralAgenda, error) {
	newGAgds := make([]NewGeneralAgenda, n)

	now := time.Now()

	diff := int(math.Min(float64(24-now.Hour()), 2))
	for i := range n {
		newGAgds[i] = NewGeneralAgenda{
			BusinessID:  bsnID,
			OpensAt:     time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.Local),
			ClosedAt:    time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+diff, 0, 0, 0, time.Local),
			Interval:    60 * 15, // 15 minutes
			WorkingDays: []Day{{2}, {4}, {6}},
		}
	}

	return newGAgds, nil
}

func TestGenerateSeedGeneralAgendas(n int, agdCore *Core, bsnID uuid.UUID, userID uuid.UUID) ([]GeneralAgenda, error) {
	newGAgds, err := TestGenerateNewGeneralAgendas(n, bsnID, userID)
	if err != nil {
		return nil, err
	}

	agds := make([]GeneralAgenda, len(newGAgds))
	for i, na := range newGAgds {
		a, err := agdCore.CreateGeneralAgenda(context.Background(), na)
		if err != nil {
			return nil, fmt.Errorf("seeding general agenda: idx: %d: %w", i, err)
		}
		agds[i] = a
	}

	return agds, nil
}

// ------------------------------------------------------------------------------------------------------------------

func TestGenerateNewDailyAgendas(n int, bsnID uuid.UUID, userID uuid.UUID) ([]NewDailyAgenda, error) {
	newDAgds := make([]NewDailyAgenda, n)

	then := time.Now().AddDate(0, 0, 2)

	diff := int(math.Min(float64(24-then.Hour()), 2))
	for i := range n {
		newDAgds[i] = NewDailyAgenda{
			BusinessID:   bsnID,
			OpensAt:      time.Date(then.Year(), then.Month(), then.Day(), then.Hour(), 0, 0, 0, time.Local),
			ClosedAt:     time.Date(then.Year(), then.Month(), then.Day(), then.Hour()+diff, 0, 0, 0, time.Local),
			Interval:     60 * 10, // 10 minutes
			Date:         time.Date(then.Year(), then.Month(), then.Day(), 0, 0, 0, 0, time.Local),
			Availability: true,
		}
	}

	return newDAgds, nil
}

func TestGenerateSeedDailyAgendas(n int, agdCore *Core, bsnID uuid.UUID, userID uuid.UUID) ([]DailyAgenda, error) {
	newDAgds, err := TestGenerateNewDailyAgendas(n, bsnID, userID)
	if err != nil {
		return nil, err
	}

	agds := make([]DailyAgenda, len(newDAgds))
	for i, na := range newDAgds {
		a, err := agdCore.CreateDailyAgenda(context.Background(), na)
		if err != nil {
			return nil, fmt.Errorf("seeding daily agenda: idx: %d: %w", i, err)
		}
		agds[i] = a
	}

	return agds, nil
}

// ---------------------------------------------------------------------------------------------

// GetWorkingDays should only be used for testing purposes.
func GetWorkingDays(val ...uint) ([]Day, error) {
	days := make([]Day, len(val))

	for i := range days {
		pd, err := ParseDay(val[i])
		if err != nil {
			return nil, err
		}
		days[i] = pd
	}

	return days, nil
}

func DurationPointer(d time.Duration) *time.Duration {
	return &d
}
