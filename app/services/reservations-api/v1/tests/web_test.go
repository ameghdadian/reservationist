package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ameghdadian/service/app/services/reservations-api/v1/cmd/all"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/appointmentgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/businessgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/usergrp"
	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/business/data/order"
	v1 "github.com/ameghdadian/service/business/web/v1"
	"github.com/ameghdadian/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
)

type seedData struct {
	users        []user.User
	businesses   []business.Business
	appointments []appointment.Appointment
}

// WebTests holds methods for each subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtest are registered.
type WebTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

func Test_Web(t *testing.T) {
	t.Parallel()

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	shutdown := make(chan os.Signal, 1)
	tests := WebTests{
		app: v1.APIMux(v1.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.V1.Auth,
			DB:       test.DB,
		}, all.Routes()),
		userToken:  test.TokenV1("user@example.com", "gophers"),
		adminToken: test.TokenV1("admin@example.com", "gophers"),
	}

	// ================================================================

	seed := func(ctx context.Context, api dbtest.CoreAPIs) (seedData, error) {
		usrs, err := api.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding users: %w", err)
		}

		bsns1, err := business.TestGenerateSeedBusinesses(1, api.Business, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding businesses: %w", err)
		}

		bsns2, err := business.TestGenerateSeedBusinesses(1, api.Business, usrs[1].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding businesses: %w", err)
		}

		var bsns []business.Business
		bsns = append(bsns, bsns1...)
		bsns = append(bsns, bsns2...)

		apts1, err := appointment.TestGenerateSeedAppointments(1, api.Appointment, usrs[0].ID, bsns1[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding appointments: %w", err)
		}

		apts2, err := appointment.TestGenerateSeedAppointments(1, api.Appointment, usrs[1].ID, bsns2[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding appointments: %w", err)
		}

		var apts []appointment.Appointment
		apts = append(apts, apts1...)
		apts = append(apts, apts2...)

		sd := seedData{
			users:        usrs,
			businesses:   bsns,
			appointments: apts,
		}

		return sd, nil
	}

	t.Log("Seeding data ...")

	sd, err := seed(context.Background(), api)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ================================================================

	t.Run("query200", tests.query200(sd))
	t.Run("queryByID200", tests.queryByID200(sd))
	t.Run("createUser200", tests.createUser200(sd))
	t.Run("createBusiness200", tests.createBusiness200(sd))
	t.Run("createAppointment200", tests.createAppointment200(sd))
}

func (wt *WebTests) query200(sd seedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []struct {
			name    string
			url     string
			resp    any
			expResp any
		}{
			{
				name: "user",
				url:  "/v1/users?page=1&rows=2&orderBy=user_id,DESC",
				resp: &response.PageDocument[usergrp.AppUser]{},
				expResp: &response.PageDocument[usergrp.AppUser]{
					Page:        1,
					RowsPerPage: 2,
					Total:       len(sd.users),
					Items:       toAppUsers(sd.users),
				},
			},
			{
				name: "business",
				url:  "/v1/businesses?page=1&rows=2&orderBy=owner_id,DESC",
				resp: &response.PageDocument[businessgrp.AppBusiness]{},
				expResp: &response.PageDocument[businessgrp.AppBusiness]{
					Page:        1,
					RowsPerPage: 2,
					Total:       len(sd.businesses),
					Items:       toAppBusinesses(sd.businesses),
				},
			},
			{
				name: "appointment",
				url:  "/v1/appointments?page=1&rows=2&orderBy=user_id,DESC",
				resp: &response.PageDocument[appointmentgrp.AppAppointment]{},
				expResp: &response.PageDocument[appointmentgrp.AppAppointment]{
					Page:        1,
					RowsPerPage: 2,
					Total:       len(sd.appointments),
					Items:       toAppAppointments(sd.appointments),
				},
			},
		}

		for _, tt := range table {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+wt.adminToken)
			wt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Errorf("%s: Should receive a status code of 200 for the response: %d", tt.name, w.Code)
				continue
			}

			if err := json.Unmarshal(w.Body.Bytes(), &tt.resp); err != nil {
				t.Errorf("Should be able to unmarshal the respones: %s", err)
			}

			diff := cmp.Diff(tt.resp, tt.expResp)
			if diff != "" {
				t.Error("Should get the expected response")
				t.Log("GOT")
				t.Logf("%#v", tt.resp)
				t.Log("EXP")
				t.Logf("%#v", tt.expResp)
				continue
			}
		}
	}
}

func (wt *WebTests) queryByID200(sd seedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []struct {
			name    string
			url     string
			resp    any
			expResp any
		}{
			{
				name:    "user",
				url:     fmt.Sprintf("/v1/users/%s", sd.users[1].ID),
				resp:    &usergrp.AppUser{},
				expResp: toAppUserPtr(sd.users[1]),
			},
			{
				name:    "business",
				url:     fmt.Sprintf("/v1/businesses/%s", sd.businesses[1].ID),
				resp:    &businessgrp.AppBusiness{},
				expResp: toAppBusinessPtr(sd.businesses[1]),
			},
			{
				name:    "appointment",
				url:     fmt.Sprintf("/v1/appointments/%s", sd.appointments[1].ID),
				resp:    &appointmentgrp.AppAppointment{},
				expResp: toAppAppointmentPtr(sd.appointments[1]),
			},
		}

		for _, tt := range table {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+wt.userToken)
			wt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Errorf("%s: Should receive a status code of 200 for the response: %d", tt.name, w.Code)
				continue
			}

			if err := json.Unmarshal(w.Body.Bytes(), &tt.resp); err != nil {
				t.Errorf("%s: Should be able to unmarshal the respones: %s", tt.name, err)
			}

			diff := cmp.Diff(tt.resp, tt.expResp)
			if diff != "" {
				t.Error("Should get the expected response")
				t.Log("GOT")
				t.Logf("%#v", tt.resp)
				t.Log("EXP")
				t.Logf("%#v", tt.expResp)
				continue
			}
		}

	}
}

func (wt *WebTests) createUser200(sd seedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []struct {
			name    string
			url     string
			input   any
			resp    any
			expResp any
		}{
			{
				name: "user",
				url:  "/v1/users",
				input: &usergrp.AppNewUser{
					Name:            "John Doe",
					Email:           "j.doe@gmail.com",
					Roles:           []string{"ADMIN"},
					PhoneNo:         "+989121928374",
					Password:        "123",
					PasswordConfirm: "123",
				},
				resp: &usergrp.AppUser{},
				expResp: &usergrp.AppUser{
					Name:    "John Doe",
					Email:   "j.doe@gmail.com",
					Roles:   []string{"ADMIN"},
					PhoneNo: "+989121928374",
					Enabled: true,
				},
			},
		}

		for _, tt := range table {
			d, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("error occurred")
			}

			r := httptest.NewRequest(http.MethodPost, tt.url, bytes.NewBuffer(d))
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+wt.adminToken)
			wt.app.ServeHTTP(w, r)

			if w.Code != http.StatusCreated {
				t.Errorf("%s: Should receive a status code of 201 for the response: %d", tt.name, w.Code)
				continue
			}

			if err := json.Unmarshal(w.Body.Bytes(), &tt.resp); err != nil {
				t.Errorf("Should be able to unmarshal the respones: %s", err)
			}

			gotResp, exists := tt.resp.(*usergrp.AppUser)
			if !exists {
				t.Fatalf("error occurred")
			}

			expResp := tt.expResp.(*usergrp.AppUser)
			expResp.ID = gotResp.ID
			expResp.DateCreated = gotResp.DateCreated
			expResp.DateUpdated = gotResp.DateUpdated

			diff := cmp.Diff(gotResp, expResp)
			if diff != "" {
				t.Error("Should get the expected response")
				t.Log("GOT")
				t.Logf("%#v", gotResp)
				t.Log("EXP")
				t.Logf("%#v", expResp)
				continue
			}
		}
	}
}

func (wt *WebTests) createBusiness200(sd seedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []struct {
			name    string
			url     string
			input   any
			resp    any
			expResp any
		}{
			{
				name: "business",
				url:  "/v1/businesses",
				input: &businessgrp.AppNewBusiness{
					Name:        "New Business",
					OwnerID:     sd.users[0].ID.String(),
					Description: "New Businesss Description",
				},
				resp: &businessgrp.AppBusiness{},
				expResp: &businessgrp.AppBusiness{
					Name:        "New Business",
					OwnerID:     sd.users[0].ID.String(),
					Description: "New Businesss Description",
				},
			},
		}

		for _, tt := range table {
			d, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("error occurred")
			}

			r := httptest.NewRequest(http.MethodPost, tt.url, bytes.NewBuffer(d))
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+wt.adminToken)
			wt.app.ServeHTTP(w, r)

			if w.Code != http.StatusCreated {
				t.Errorf("%s: Should receive a status code of 201 for the response: %d", tt.name, w.Code)
				continue
			}

			if err := json.Unmarshal(w.Body.Bytes(), &tt.resp); err != nil {
				t.Errorf("Should be able to unmarshal the respones: %s", err)
			}

			gotResp, exists := tt.resp.(*businessgrp.AppBusiness)
			if !exists {
				t.Fatalf("error occurred")
			}

			expResp := tt.expResp.(*businessgrp.AppBusiness)
			expResp.ID = gotResp.ID
			expResp.DateCreated = gotResp.DateCreated
			expResp.DateUpdated = gotResp.DateUpdated

			diff := cmp.Diff(gotResp, expResp)
			if diff != "" {
				t.Error("Should get the expected response")
				t.Log("GOT")
				t.Logf("%#v", gotResp)
				t.Log("EXP")
				t.Logf("%#v", expResp)
				continue
			}
		}
	}
}

func (wt *WebTests) createAppointment200(sd seedData) func(t *testing.T) {
	inTwoHrs := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)

	return func(t *testing.T) {
		table := []struct {
			name    string
			url     string
			input   any
			resp    any
			expResp any
		}{
			{
				name: "appointment",
				url:  "/v1/appointments",
				input: &appointmentgrp.AppNewAppointment{
					BusinessID:  sd.businesses[0].ID.String(),
					UserID:      sd.users[0].ID.String(),
					Status:      appointment.StatusScheduled.Status(),
					ScheduledOn: inTwoHrs,
				},
				resp: &appointmentgrp.AppAppointment{},
				expResp: &appointmentgrp.AppAppointment{
					BusinessID:  sd.businesses[0].ID.String(),
					UserID:      sd.users[0].ID.String(),
					Status:      appointment.StatusScheduled.Status(),
					ScheduledOn: inTwoHrs,
				},
			},
		}

		t.Log(table[0].input)

		for _, tt := range table {
			d, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("error occurred")
			}

			r := httptest.NewRequest(http.MethodPost, tt.url, bytes.NewBuffer(d))
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+wt.adminToken)
			wt.app.ServeHTTP(w, r)

			if w.Code != http.StatusCreated {
				t.Errorf("%s: Should receive a status code of 201 for the response: %d", tt.name, w.Code)
				continue
			}

			if err := json.Unmarshal(w.Body.Bytes(), &tt.resp); err != nil {
				t.Errorf("Should be able to unmarshal the respones: %s", err)
			}

			gotResp, exists := tt.resp.(*appointmentgrp.AppAppointment)
			if !exists {
				t.Fatalf("error occurred")
			}

			expResp := tt.expResp.(*appointmentgrp.AppAppointment)
			expResp.ID = gotResp.ID
			expResp.DateCreated = gotResp.DateCreated
			expResp.DateUpdated = gotResp.DateUpdated

			diff := cmp.Diff(gotResp, expResp)
			if diff != "" {
				t.Error("Should get the expected response")
				t.Log("GOT")
				t.Logf("%#v", gotResp)
				t.Log("EXP")
				t.Logf("%#v", expResp)
				continue
			}
		}
	}
}
