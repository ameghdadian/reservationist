package agenda_test

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

type seedData struct {
	usrs  []user.User
	bsns  []business.Business
	gAgds []agenda.GeneralAgenda
	dAgds []agenda.DailyAgenda
}

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	fmt.Println("Starting a new database")
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_Agenda(t *testing.T) {
	t.Run("crud", crud)
}

func crud(t *testing.T) {
	seed := func(ctx context.Context, agdCore *agenda.Core, bsnCore *business.Core, usrCore *user.Core) (seedData, error) {
		var filter user.QueryFilter
		filter.WithName("User Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return seedData{}, err
		}

		bsns, err := business.TestGenerateSeedBusinesses(2, bsnCore, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding bsns: %w", err)
		}

		gAgds, err := agenda.TestGenerateSeedGeneralAgendas(1, agdCore, bsns[0].ID, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding gAgds: %w", err)
		}

		dAgds, err := agenda.TestGenerateSeedDailyAgendas(1, agdCore, bsns[0].ID, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding dAgds: %w", err)
		}

		return seedData{
			usrs:  usrs,
			bsns:  bsns,
			gAgds: gAgds,
			dAgds: dAgds,
		}, nil
	}

	// ----------------------------------------------------------------------------------------------------------------

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Go seeding ...")

	sd, err := seed(ctx, api.Agenda, api.Business, api.User)
	if err != nil {
		t.Fatalf("Seeding error: %s\n", err)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// 												General Agenda
	// ----------------------------------------------------------------------------------------------------------------

	// ----------------------------------------------------------------------------------------------------------------
	// Count

	var filter agenda.GAQueryFilter
	filter.WithBusinessID(sd.bsns[0].ID)
	n, err := api.Agenda.CountGeneralAgenda(ctx, filter)
	if err != nil {
		t.Fatal("Should be able to count general agendas")
	}

	if n != 1 {
		t.Error("Should have the correct number of general agendas")
		t.Errorf("GOT: %d\n", n)
		t.Errorf("EXP: %d\n", 1)
	}

	filter.WithBusinessID(uuid.New())
	n, err = api.Agenda.CountGeneralAgenda(ctx, filter)
	if err != nil {
		t.Errorf("GOT: %s\n", err)
		t.Error("EXP: NO ERROR RETURNED")
	}
	if n != 0 {
		t.Error("Should have the correct number of general agendas")
		t.Errorf("GOT: %d\n", n)
		t.Errorf("EXP: %d\n", 0)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// QueryByID

	saved, err := api.Agenda.QueryGeneralAgenda(ctx, agenda.GAQueryFilter{ID: &sd.gAgds[0].ID})
	if err != nil {
		t.Fatalf("Should be able to retrieve general agenda by ID: %s", err)
	}

	if saved.DateCreated.UnixMilli() != sd.gAgds[0].DateCreated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateCreated)
		t.Logf("EXP: %v", sd.gAgds[0].DateCreated)
		t.Errorf("Should get back the same date created")
	}
	if saved.DateUpdated.UnixMilli() != sd.gAgds[0].DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateUpdated)
		t.Logf("EXP: %v", sd.gAgds[0].DateUpdated)
		t.Errorf("Should get back the same date updated")
	}

	sd.gAgds[0].DateCreated = time.Time{}
	sd.gAgds[0].DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(sd.gAgds[0], saved); diff != "" {
		t.Errorf("Should get back the same general agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// Create

	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now()
	wd, err := agenda.GetWorkingDays(0, 2, 3)
	if err != nil {
		t.Fatal("Should be able to call GetWorkingDays: s", err)
	}

	nga := agenda.NewGeneralAgenda{
		BusinessID:  sd.bsns[1].ID,
		OpensAt:     time.Date(now.Year(), now.Month(), now.Day(), 14, 10, 0, 0, loc),
		ClosedAt:    time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc),
		Interval:    60 * time.Second * 20,
		WorkingDays: wd,
	}

	gagd, err := api.Agenda.CreateGeneralAgenda(ctx, nga)
	if err != nil {
		t.Fatalf("Should be able to create a general agenda: %s", err)
	}

	saved, err = api.Agenda.QueryGeneralAgenda(ctx, agenda.GAQueryFilter{ID: &gagd.ID})
	if err != nil {
		t.Fatalf("Should be able to query general agenda by id: %s", err)
	}

	if saved.DateCreated.UnixMilli() != gagd.DateCreated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateCreated)
		t.Logf("EXP: %v", gagd.DateCreated)
		t.Errorf("Should get back the same date created")
	}
	if saved.DateUpdated.UnixMilli() != gagd.DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateUpdated)
		t.Logf("EXP: %v", gagd.DateUpdated)
		t.Errorf("Should get back the same date updated")
	}

	gagd.DateCreated = time.Time{}
	gagd.DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(gagd, saved); diff != "" {
		t.Errorf("Should get back the same general agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// QueryByBusinessID

	// ga, err := api.Agenda.QueryGeneralAgenda(ctx, agenda.GAQueryFilter{BusinessID: &sd.bsns[0].ID})
	ga, err := api.Agenda.QueryGeneralAgenda(ctx, agenda.GAQueryFilter{BusinesesID: &sd.bsns[0].ID})
	if err != nil {
		t.Fatalf("Should be able to query general agenda by business id: %s", err)
	}

	ga.DateCreated = time.Time{}
	ga.DateUpdated = time.Time{}

	if diff := cmp.Diff(ga, sd.gAgds[0]); diff != "" {
		t.Errorf("Should get back the same general agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// Update

	wd, _ = agenda.GetWorkingDays(0, 1, 3)
	uga := agenda.UpdateGeneralAgenda{WorkingDays: wd}
	agd, err := api.Agenda.UpdateGenralAgenda(ctx, sd.gAgds[0], uga)
	if err != nil {
		t.Fatalf("Should be able to update general agenda: %s", err)
	}

	saved, err = api.Agenda.QueryGeneralAgenda(ctx, agenda.GAQueryFilter{ID: &sd.gAgds[0].ID})
	if err != nil {
		t.Fatalf("Should be able to query general agenda by id: %s", err)
	}

	if saved.DateUpdated.UnixMilli() != agd.DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateUpdated)
		t.Logf("EXP: %v", agd.DateUpdated)
		t.Errorf("Should get back the same date updated")
	}

	agd.DateCreated = time.Time{}
	saved.DateCreated = time.Time{}
	agd.DateUpdated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(agd, saved); diff != "" {
		t.Errorf("Should get back the same general agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// Delete

	err = api.Agenda.DeleteGeneralAgenda(ctx, sd.gAgds[0])
	if err != nil {
		t.Fatalf("Should be able to delete general agenda")
	}

	agd, err = api.Agenda.QueryGeneralAgenda(ctx, agenda.GAQueryFilter{ID: &sd.gAgds[0].ID})
	if err != nil {
		if !errors.Is(err, agenda.ErrNotFound) {
			t.Fatalf("Should be deleted by now: %s", err)
		}
	}

	// ----------------------------------------------------------------------------------------------------------------
	// 												Daily Agenda
	// ----------------------------------------------------------------------------------------------------------------

	// ----------------------------------------------------------------------------------------------------------------
	// Count

	var dfilter agenda.DAQueryFilter
	dfilter.WithDays(3)
	n, err = api.Agenda.CountDailyAgenda(ctx, dfilter)
	if err != nil {
		t.Fatalf("Should be able to count daily agendas: %s\n", err)
	}

	if n != 1 {
		t.Error("Should have the correct number of daily agendas")
		t.Errorf("GOT: %d\n", n)
		t.Errorf("EXP: %d\n", 1)
	}

	dfilter.WithDays(1)
	n, err = api.Agenda.CountDailyAgenda(ctx, dfilter)
	if err != nil {
		t.Errorf("GOT: %s\n", err)
		t.Error("EXP: NO ERROR RETURNED")
	}
	if n != 0 {
		t.Error("Should have the correct number of daily agendas")
		t.Errorf("GOT: %d\n", n)
		t.Errorf("EXP: %d\n", 0)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// QueryByID

	daSaved, err := api.Agenda.QueryDailyAgendaByID(ctx, sd.dAgds[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve daily agenda by ID: %s", err)
	}

	if daSaved.DateCreated.UnixMilli() != sd.dAgds[0].DateCreated.UnixMilli() {
		t.Logf("GOT: %v", daSaved.DateCreated)
		t.Logf("EXP: %v", sd.dAgds[0].DateCreated)
		t.Errorf("Should get back the same date created")
	}
	if daSaved.DateUpdated.UnixMilli() != sd.dAgds[0].DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", daSaved.DateUpdated)
		t.Logf("EXP: %v", sd.dAgds[0].DateUpdated)
		t.Errorf("Should get back the same date updated")
	}

	sd.dAgds[0].DateCreated = time.Time{}
	sd.dAgds[0].DateUpdated = time.Time{}
	daSaved.DateCreated = time.Time{}
	daSaved.DateUpdated = time.Time{}

	if diff := cmp.Diff(sd.dAgds[0], daSaved); diff != "" {
		t.Errorf("Should get backthe same daily agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// Create

	loc, _ = time.LoadLocation("America/New_York")
	now = time.Now()
	tommorow := now.AddDate(0, 0, 1)

	nda := agenda.NewDailyAgenda{
		BusinessID: sd.bsns[1].ID,
		OpensAt:    time.Date(now.Year(), now.Month(), now.Day(), 14, 10, 0, 0, loc),
		ClosedAt:   time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc),
		Interval:   60 * time.Second * 20,
		// Dates are stored in DB based on UTC timezone, and returned in time.Local. We're doing the same here to mimic that behavior.
		Date:         time.Date(tommorow.Year(), tommorow.Month(), tommorow.Day(), 0, 0, 0, 0, time.UTC).In(time.Local),
		Availability: true,
	}

	dagd, err := api.Agenda.CreateDailyAgenda(ctx, nda)
	if err != nil {
		t.Fatalf("Should be able to create a daily agenda: %s", err)
	}

	daSaved, err = api.Agenda.QueryDailyAgendaByID(ctx, dagd.ID)
	if err != nil {
		t.Fatalf("Should be able to query daily agenda by id: %s", err)
	}

	if daSaved.DateCreated.UnixMilli() != dagd.DateCreated.UnixMilli() {
		t.Logf("GOT: %v", daSaved.DateCreated)
		t.Logf("EXP: %v", dagd.DateCreated)
		t.Errorf("Should get back the same date created")
	}
	if daSaved.DateUpdated.UnixMilli() != dagd.DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", daSaved.DateUpdated)
		t.Logf("EXP: %v", dagd.DateUpdated)
		t.Errorf("Should get back the same date updated")
	}

	dagd.DateCreated = time.Time{}
	dagd.DateUpdated = time.Time{}
	daSaved.DateCreated = time.Time{}
	daSaved.DateUpdated = time.Time{}

	if diff := cmp.Diff(dagd, daSaved); diff != "" {
		t.Errorf("Should get back the same daily agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// QueryByBusinessID

	da, err := api.Agenda.QueryDailyAgendaByBusinessID(ctx, sd.bsns[0].ID)
	if err != nil {
		t.Fatalf("Should be able to query daily agenda by business id: %s", err)
	}

	da[0].DateCreated = time.Time{}
	da[0].DateUpdated = time.Time{}

	if diff := cmp.Diff(da[0], sd.dAgds[0]); diff != "" {
		t.Errorf("Should get back the same daily agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// Update

	dura := time.Duration(1) * time.Hour
	uda := agenda.UpdateDailyAgenda{Interval: agenda.DurationPointer(dura)}
	dagd, err = api.Agenda.UpdateDailyAgenda(ctx, sd.dAgds[0], uda)
	if err != nil {
		t.Fatalf("Should be able to update daily agenda: %s", err)
	}

	daSaved, err = api.Agenda.QueryDailyAgendaByID(ctx, sd.dAgds[0].ID)
	if err != nil {
		t.Fatalf("Should be able to query daily agenda by id: %s", err)
	}

	if daSaved.DateUpdated.UnixMilli() != dagd.DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", daSaved.DateUpdated)
		t.Logf("EXP: %v", dagd.DateUpdated)
		t.Errorf("Should get back the same date updated")
	}

	dagd.DateCreated = time.Time{}
	daSaved.DateCreated = time.Time{}
	dagd.DateUpdated = time.Time{}
	daSaved.DateUpdated = time.Time{}

	if diff := cmp.Diff(dagd, daSaved); diff != "" {
		t.Errorf("Should get back the same general agenda, diff:\n%s", diff)
	}

	// ----------------------------------------------------------------------------------------------------------------
	// Delete

	err = api.Agenda.DeleteDailyAgenda(ctx, sd.dAgds[0])
	if err != nil {
		t.Fatalf("Should be able to delete daily agenda")
	}

	dagd, err = api.Agenda.QueryDailyAgendaByID(ctx, sd.dAgds[0].ID)
	if err != nil {
		if !errors.Is(err, agenda.ErrNotFound) {
			t.Fatalf("Should be deleted by now: %s", err)
		}
	}
}
