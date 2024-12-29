package business_test

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/business/data/page"
	"github.com/ameghdadian/service/business/data/redistest"
	"github.com/ameghdadian/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

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

	fmt.Println("Starting a new redis")
	rc, err = redistest.StartRedis()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer redistest.StopRedis(rc)
	m.Run()
}

func Test_Business(t *testing.T) {
	t.Run("crud", crud)
}

func crud(t *testing.T) {
	seed := func(ctx context.Context, bsnCore *business.Core, usrCore *user.Core) ([]business.Business, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		pagination := page.MustParse("1", "1")
		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, pagination)
		if err != nil {
			return nil, fmt.Errorf("seeding users: %w", err)
		}

		bsns, err := business.TestGenerateSeedBusinesses(1, bsnCore, usrs[0].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding businesses: %w", err)
		}

		return bsns, err
	}

	// -------------------------------------------------------------------

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

	bsns, err := seed(ctx, api.Business, api.User)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------
	// Count

	n, err := api.Business.Count(ctx, business.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to count business")
	}

	if n != 1 {
		t.Error("Should have the correct number of businesses")
		t.Errorf("GOT: %d\n", n)
		t.Errorf("EXP: %d\n", 1)
	}

	// -------------------------------------------------------------------
	// QueryByID

	saved, err := api.Business.QueryByID(ctx, bsns[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve business by ID: %s", err)
	}

	if bsns[0].DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateCreated)
		t.Logf("EXP: %v", bsns[0].DateCreated)
		t.Logf("DIFF: %v", saved.DateCreated.Sub(bsns[0].DateCreated))
		t.Errorf("Should get back the same date created")
	}

	if bsns[0].DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateUpdated)
		t.Logf("EXP: %v", bsns[0].DateUpdated)
		t.Logf("DIFF: %v", saved.DateUpdated.Sub(bsns[0].DateUpdated))
		t.Errorf("Should get back the same date updated")
	}

	bsns[0].DateCreated = time.Time{}
	bsns[0].DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(bsns[0], saved); diff != "" {
		t.Errorf("Should get back the same product, diff:\n%s", diff)
	}

	// -------------------------------------------------------------------
	// Create Business

	nb := business.NewBusiness{
		OwnerID: bsns[0].OwnerID,
		Name:    "New business",
		Desc:    "A prosperous one!",
	}
	b, err := api.Business.Create(ctx, nb)
	if err != nil {
		t.Fatalf("Should be able to create a business: %s", err)
	}

	if b.OwnerID == uuid.Nil {
		t.Error("Should have a valid business owner id")
	}
	if b.ID == uuid.Nil {
		t.Error("Should have a valid business id")
	}
	if b.Name != nb.Name {
		t.Error("Should have the correct business name.")
		t.Errorf("GOT: %s\n", b.Name)
		t.Errorf("EXP: %s\n", nb.Name)
	}
	if b.Desc != nb.Desc {
		t.Error("Should have the correct business description.")
		t.Errorf("GOT: %s\n", b.Desc)
		t.Errorf("EXP: %s\n", nb.Desc)
	}
	if time.Now().UnixMilli()-b.DateCreated.UnixMilli() > time.Second.Milliseconds() {
		t.Error("Should be created just recently.")
		t.Errorf("GOT: %s\n", b.DateCreated)

	}
	if time.Now().UnixMilli()-b.DateUpdated.UnixMilli() > time.Second.Milliseconds() {
		t.Error("Should be updated just recently.")
		t.Errorf("GOT: %s\n", b.DateUpdated)
	}

	// -------------------------------------------------------------------
	// QueryByOwnerID

	savedBsns, err := api.Business.QueryByOwnerID(ctx, bsns[0].OwnerID)
	if err != nil {
		t.Fatalf("Should be able to query business by Owner ID.")
	}
	// Up to now, we should have two businesses associated with this Owner ID (considering one we just created)
	if len(savedBsns) != 2 {
		t.Errorf("Should have 2 businesses: OwnerID[%s]\n", bsns[0].OwnerID)
		t.Errorf("GOT: %d\n", len(savedBsns))
		t.Errorf("EXP: %d\n", 2)
	}

	// -------------------------------------------------------------------
	// Update

	ub := business.UpdateBusiness{Name: dbtest.StringPointer("New updated business")}
	b, err = api.Business.Update(ctx, bsns[0], ub)
	if err != nil {
		t.Fatalf("Should be able to update business.")
	}

	if b.Name != *ub.Name {
		t.Error("Should have the new updated name")
		t.Errorf("EXP: %s\n", *ub.Name)
		t.Errorf("GOT: %s\n", b.Name)
	}

	// -------------------------------------------------------------------
	// Delete

	err = api.Business.Delete(ctx, bsns[0])
	if err != nil {
		t.Fatalf("Should be able to delete business")
	}

	b, err = api.Business.QueryByID(ctx, bsns[0].ID)
	if err != nil {
		if !errors.Is(err, business.ErrNotFound) {
			t.Fatalf("Should be deleted by now.")
		}
	}

}
