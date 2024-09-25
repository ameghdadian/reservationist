package user_test

import (
	"context"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/foundation/docker"
	"github.com/google/uuid"
)

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

func Test_User(t *testing.T) {
	t.Run("crud", crud)
}

// =======================================================

func crud(t *testing.T) {
	// Used to seed database with whatever seed data that we might need
	seed := func(ctx context.Context, usrCore *user.Core) ([]user.User, error) {
		return []user.User{}, nil
	}

	// ===================================================

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

	_, err := seed(ctx, api.User)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ===================================================

	pn, err := user.ParsePhoneNumber("+989129129129")
	if err != nil {
		t.Fatalf("Should be able to parse phone number: %s.", err)
	}
	email, err := mail.ParseAddress("jd@gmail.com")
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
}
