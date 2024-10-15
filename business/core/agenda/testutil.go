package agenda

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func TestGenerateNewGeneralAgendas(n int, bsnID uuid.UUID, userID uuid.UUID) ([]NewGeneralAgenda, error) {
	newGAgds := make([]NewGeneralAgenda, n)

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}
	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	for i := range n {
		newGAgds[i] = NewGeneralAgenda{
			BusinessID:  bsnID,
			OpensAt:     midnight.Add(9 * time.Hour),  // Business opens at 9 O'clock in NewYork timezone
			ClosedAt:    midnight.Add(17 * time.Hour), // Business is closed at 17 o'clock in NewYork timezone
			Interval:    60 * time.Second * 15,        // 15 minutes
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

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}
	now := time.Now().In(loc)
	then := now.AddDate(0, 0, 2)
	thenMidnight := time.Date(then.Year(), then.Month(), then.Day(), 0, 0, 0, 0, loc)

	for i := range n {
		newDAgds[i] = NewDailyAgenda{
			BusinessID: bsnID,
			OpensAt:    thenMidnight.Add(9 * time.Hour),  // Business opens at 9 O'clock in NewYork timezone
			ClosedAt:   thenMidnight.Add(12 * time.Hour), // Business is closed at 12 o'clock in NewYork timezone
			Interval:   60 * time.Second * 10,            // 15 minutes
			// Dates are stored in DB based on UTC timezone, and returned in time.Local. We're doing the same here to mimic that behavior.
			Date:         time.Date(thenMidnight.Year(), thenMidnight.Month(), thenMidnight.Day(), 0, 0, 0, 0, time.UTC).In(time.Local),
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

// GetWorkingTest should only be used for testing purposes.
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
