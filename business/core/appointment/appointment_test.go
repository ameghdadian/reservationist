package appointment_test

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/business/data/redistest"
	"github.com/ameghdadian/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

type seedData struct {
	apts []appointment.Appointment
	usrs []user.User
	bsns []business.Business
}

var c *docker.Container
var rc *docker.Container

func TestMain(m *testing.M) {
	var err error
	fmt.Println("Starting a new database")
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	rc, err = redistest.StartRedis()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer redistest.StopRedis(rc)

	m.Run()
}

func Test_Appointment(t *testing.T) {
	t.Run("crud", crud)
}

func crud(t *testing.T) {
	seed := func(ctx context.Context, aptCore *appointment.Core, usrCore *user.Core, bsnCore *business.Core) (seedData, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding users: %w", err)
		}

		bsns, err := business.TestGenerateSeedBusinesses(1, bsnCore, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding bsns: %w", err)
		}

		apts, err := appointment.TestGenerateSeedAppointments(1, aptCore, usrs[0].ID, bsns[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding apts: %w", err)
		}

		return seedData{
			apts: apts,
			usrs: usrs,
			bsns: bsns,
		}, nil
	}

	// -----------------------------------------------------------------------------------------------------

	test := dbtest.NewTest(t, c, rc)
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

	sd, err := seed(ctx, api.Appointment, api.User, api.Business)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -----------------------------------------------------------------------------------------------------
	// Count

	var filter appointment.QueryFilter
	filter.WithUserID(sd.usrs[0].ID)
	n, err := api.Appointment.Count(ctx, filter)
	if err != nil {
		t.Fatal("Should be able to count appointments")
	}

	if n != 1 {
		t.Error("Should have the correct number of appointments")
		t.Errorf("GOT: %d\n", n)
		t.Errorf("EXP: %d\n", 1)
	}

	// -----------------------------------------------------------------------------------------------------
	// QueryByID

	saved, err := api.Appointment.QueryByID(ctx, sd.apts[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve appointment by ID: %s", err)
	}

	if saved.DateCreated.UnixMilli() != sd.apts[0].DateCreated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateCreated)
		t.Logf("EXP: %v", sd.apts[0].DateCreated)
		t.Logf("DIFF: %v", saved.DateCreated.Sub(sd.apts[0].DateCreated))
		t.Errorf("Should get back the same date created")
	}
	if saved.DateUpdated.UnixMilli() != sd.apts[0].DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateUpdated)
		t.Logf("EXP: %v", sd.apts[0].DateUpdated)
		t.Logf("DIFF: %v", saved.DateUpdated.Sub(sd.apts[0].DateUpdated))
		t.Errorf("Should get back the same date updated")
	}
	if saved.ScheduledOn.UnixMilli() != sd.apts[0].ScheduledOn.UnixMilli() {
		t.Logf("GOT: %v", saved.ScheduledOn)
		t.Logf("EXP: %v", sd.apts[0].ScheduledOn)
		t.Logf("DIFF: %v", saved.ScheduledOn.Sub(sd.apts[0].ScheduledOn))
		t.Errorf("Should get back the same scheduled on date")
	}

	sd.apts[0].DateCreated = time.Time{}
	sd.apts[0].DateUpdated = time.Time{}
	sd.apts[0].ScheduledOn = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}
	saved.ScheduledOn = time.Time{}

	if diff := cmp.Diff(sd.apts[0], saved); diff != "" {
		t.Errorf("Should get back the same appointment, diff:\n%s", diff)
	}

	// -----------------------------------------------------------------------------------------------------
	// Create

	na := appointment.NewAppointment{
		BusinessID:  sd.bsns[0].ID,
		UserID:      sd.usrs[0].ID,
		Status:      appointment.StatusScheduled,
		ScheduledOn: time.Now().Add(2 * time.Hour),
	}
	a, err := api.Appointment.Create(ctx, na)
	if err != nil {
		t.Fatalf("Should be able to create a appointment: %s", err)
	}

	if a.ID == uuid.Nil {
		t.Error("Should have a valid appointment id")
		t.Errorf("GOT: %v\n", a.ID)
		t.Errorf("EXP: %v\n", na.BusinessID.ID())
	}
	if a.BusinessID == uuid.Nil {
		t.Error("Should have a valid business owner id")
		t.Errorf("GOT: %v\n", a.BusinessID)
		t.Errorf("EXP: %v\n", na.BusinessID)
	}
	if a.UserID == uuid.Nil {
		t.Error("Should have a valid user id")
		t.Errorf("GOT: %v\n", a.UserID)
		t.Errorf("EXP: %v\n", na.UserID)
	}
	if a.Status != na.Status {
		t.Error("Should have the correct status.")
		t.Errorf("GOT: %v\n", a.Status)
		t.Errorf("EXP: %v\n", na.Status)
	}
	if a.ScheduledOn != na.ScheduledOn {
		t.Error("Should have the correct scheduled on datetime.")
		t.Errorf("GOT: %s\n", a.ScheduledOn)
		t.Errorf("EXP: %s\n", na.ScheduledOn)
	}
	if time.Now().UnixMilli()-a.DateCreated.UnixMilli() > time.Second.Milliseconds() {
		t.Error("Should be created just recently.")
		t.Errorf("GOT: %s\n", a.DateCreated)

	}
	if time.Now().UnixMilli()-a.DateUpdated.UnixMilli() > time.Second.Milliseconds() {
		t.Error("Should be updated just recently.")
		t.Errorf("GOT: %s\n", a.DateUpdated)
	}

	// -------------------------------------------------------------------
	// QueryByOwnerID

	savedApts, err := api.Appointment.QueryByUserID(ctx, sd.usrs[0].ID)
	if err != nil {
		t.Fatalf("Should be able to query business by Owner ID.")
	}
	// Up to now, we should have two appointments associated with this User ID (considering one we just created)
	if len(savedApts) != 2 {
		t.Errorf("Should have 2 appointments: UserID[%s]\n", sd.apts[0].UserID)
		t.Errorf("GOT: %d\n", len(savedApts))
		t.Errorf("EXP: %d\n", 2)
	}

	// -------------------------------------------------------------------
	// Update

	// Restore back scheduled on time after resetting it in QueryByID test suite
	sd.apts[0].ScheduledOn = time.Now().Add(2 * time.Hour)
	ua := appointment.UpdateAppointment{Status: &appointment.StatusCancelled}
	apt1, err := api.Appointment.Update(ctx, sd.apts[0], ua)
	if err != nil {
		t.Fatalf("Should be able to update appointment: %s", err)
	}

	if apt1.Status != *ua.Status {
		t.Error("Should have the new status")
		t.Errorf("EXP: %v\n", *ua.Status)
		t.Errorf("GOT: %v\n", apt1.Status)
	}

	// -------------------------------------------------------------------
	// Delete

	err = api.Appointment.Delete(ctx, sd.apts[0])
	if err != nil {
		t.Fatalf("Should be able to delete appointment")
	}

	a, err = api.Appointment.QueryByID(ctx, sd.apts[0].ID)
	if err != nil {
		if !errors.Is(err, appointment.ErrNotFound) {
			t.Fatalf("Should be deleted by now.")
		}
	}
}
