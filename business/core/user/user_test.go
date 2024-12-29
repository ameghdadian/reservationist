package user_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/business/data/order"
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

func Test_User(t *testing.T) {
	t.Run("crud", crud)
}

// =======================================================

func crud(t *testing.T) {
	// Used to seed database with whatever seed data that we might need
	seed := func(ctx context.Context, usrCore *user.Core) ([]user.User, error) {
		usrs, err := usrCore.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, page.MustParse("1", "1"))
		if err != nil {
			return nil, fmt.Errorf("seeding users: %w", err)
		}
		return usrs, nil
	}

	allUsersSeed := func(ctx context.Context, usrCore *user.Core) ([]user.User, error) {
		usrs, err := usrCore.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, page.MustParse("1", "2"))
		if err != nil {
			return nil, fmt.Errorf("seeding users: %w", err)
		}
		return usrs, nil
	}

	// ===================================================

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

	usrs, err := seed(ctx, api.User)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	allUsrs, err := allUsersSeed(ctx, api.User)
	if err != nil {
		t.Fatalf("Seeding error: %s.", err)
	}

	// ===================================================

	userID := []uuid.UUID{allUsrs[0].ID, allUsrs[1].ID}
	savedUsrs, err := api.User.QueryByIDs(ctx, userID)
	if err != nil {
		t.Fatalf("Should be able to retrieve users by their IDs: %s", err)
	}

	if len(userID) != len(savedUsrs) {
		t.Fatalf("Should have the same number of users for seed users and db users")
	}

	if diff := cmp.Diff(allUsrs, savedUsrs); diff != "" {
		t.Fatalf("Should get back the same user. diff:\n%s", diff)
	}

	// ===================================================

	saved, err := api.User.QueryByID(ctx, usrs[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve user by ID: %s", err)
	}

	if usrs[0].DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateCreated)
		t.Logf("EXP: %v", usrs[0].DateCreated)
		t.Logf("diff: %v", saved.DateCreated.Sub(usrs[0].DateCreated))
		t.Errorf("Should get back the same created")
	}

	if usrs[0].DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
		t.Logf("GOT: %v", saved.DateUpdated)
		t.Logf("EXP: %v", usrs[0].DateUpdated)
		t.Logf("diff: %v", saved.DateUpdated.Sub(usrs[0].DateUpdated))
		t.Errorf("Should get back the same updated")
	}

	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}
	usrs[0].DateCreated = time.Time{}
	usrs[0].DateUpdated = time.Time{}

	if diff := cmp.Diff(usrs[0], saved); diff != "" {
		t.Fatalf("Should get back the same user. diff: \n%s", diff)
	}

	// ===================================================

	email, err := mail.ParseAddress("test@gmail.com")
	if err != nil {
		t.Fatalf("Should be able to parse email: %s", err)
	}

	upd := user.UpdateUser{
		Name:  dbtest.StringPointer("Test User"),
		Email: email,
	}

	if _, err := api.User.Update(ctx, usrs[0], upd); err != nil {
		t.Fatalf("Should be able to update user: %s.", err)
	}

	saved, err = api.User.QueryByEmail(ctx, *upd.Email)
	if err != nil {
		t.Fatalf("Should be able to retrieve user by Email: %s.", err)
	}

	diff := usrs[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Errorf("Should have a larger DateUpdated : sav %v, usr %v, dif %v", saved.DateUpdated, usrs[0].DateUpdated, diff)
	}

	if saved.Name != *upd.Name {
		t.Logf("GOT: %v", saved.Name)
		t.Logf("EXP: %v", *upd.Name)
		t.Errorf("Should be able to see updates to Name")
	}

	if saved.Email != *upd.Email {
		t.Logf("GOT: %v", saved.Email)
		t.Logf("EXP: %v", *upd.Email)
		t.Errorf("Should be able to see updates to Email")
	}

	if err := api.User.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete user: %s.", err)
	}

	_, err = api.User.QueryByID(ctx, saved.ID)
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve user: %s.", err)
	}

	// ===================================================

	pn, err := user.ParsePhoneNumber("+989129129129")
	if err != nil {
		t.Fatalf("Should be able to parse phone number: %s.", err)
	}
	email, err = mail.ParseAddress("jd@gmail.com")
	if err != nil {
		t.Fatalf("Shoud be able to parse email: %s.", err)
	}
	nu := user.NewUser{
		Name:            "John Doe",
		Email:           *email,
		Roles:           []user.Role{user.RoleAdmin},
		PhoneNo:         pn,
		Password:        "123",
		PasswordConfirm: "123",
	}

	usr, err := api.User.Create(context.Background(), nu)
	if err != nil {
		t.Fatalf("Should be able to create user: %s.", err)
	}

	if usr.ID == uuid.Nil {
		t.Error("Should have a valid ID.")
	}

	if usr.Name != nu.Name {
		t.Error("Should have the correct name.")
		t.Errorf("GOT: %s\n", usr.Name)
		t.Errorf("EXP: %s\n", nu.Name)
	}

	// ===================================================

	// QueryByIDs
}
