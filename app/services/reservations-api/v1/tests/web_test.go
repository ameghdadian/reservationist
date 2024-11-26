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
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/agendagrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/appointmentgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/businessgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/usergrp"
	"github.com/ameghdadian/service/business/core/agenda"
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
	users          []user.User
	businesses     []business.Business
	appointments   []appointment.Appointment
	generalAgendas []agenda.GeneralAgenda
	dailyAgendas   []agenda.DailyAgenda
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

	test := dbtest.NewTest(t, c, rc)
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
			Shutdown:      shutdown,
			Log:           test.Log,
			Auth:          test.V1.Auth,
			DB:            test.DB,
			TaskClient:    test.TaskClient,
			TaskInspector: test.TaskInspector,
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

		bsns2, err := business.TestGenerateSeedBusinesses(1, api.Business, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding businesses: %w", err)
		}

		bsns3, err := business.TestGenerateSeedBusinesses(1, api.Business, usrs[1].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding businesses: %w", err)
		}

		var bsns []business.Business
		bsns = append(bsns, bsns1...)
		bsns = append(bsns, bsns2...)
		bsns = append(bsns, bsns3...)

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

		gagd, err := agenda.TestGenerateSeedGeneralAgendas(1, api.Agenda, bsns[0].ID, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding general agenda: %w", err)
		}
		dagd, err := agenda.TestGenerateSeedDailyAgendas(1, api.Agenda, bsns[0].ID, usrs[0].ID)
		if err != nil {
			return seedData{}, fmt.Errorf("seeding daily agenda: %w", err)
		}

		sd := seedData{
			users:          usrs,
			businesses:     bsns,
			appointments:   apts,
			generalAgendas: gagd,
			dailyAgendas:   dagd,
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
	t.Run("createGeneralAgenda200", tests.createGeneralAgenda200(sd))
	// t.Run("createDailyAgenda200", tests.createDailyAgenda200(sd))
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
				url:  "/v1/businesses?page=1&rows=3&orderBy=owner_id,DESC",
				resp: &response.PageDocument[businessgrp.AppBusiness]{},
				expResp: &response.PageDocument[businessgrp.AppBusiness]{
					Page:        1,
					RowsPerPage: 3,
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
			{
				name: "general_agenda",
				url:  "/v1/agendas/general?page=1&rows=1&orderBy=id,DESC",
				resp: &response.PageDocument[agendagrp.AppGeneralAgenda]{},
				expResp: &response.PageDocument[agendagrp.AppGeneralAgenda]{
					Page:        1,
					RowsPerPage: 1,
					Total:       len(sd.generalAgendas),
					Items:       toAppGeneralAgendas(sd.generalAgendas),
				},
			},
			{
				name: "daily_agenda",
				url:  "/v1/agendas/daily?page=1&rows=1&orderBy=id,DESC",
				resp: &response.PageDocument[agendagrp.AppDailyAgenda]{},
				expResp: &response.PageDocument[agendagrp.AppDailyAgenda]{
					Page:        1,
					RowsPerPage: 1,
					Total:       len(sd.dailyAgendas),
					Items:       toAppDailyAgendas(sd.dailyAgendas),
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
				url:     fmt.Sprintf("/v1/businesses/%s", sd.businesses[2].ID),
				resp:    &businessgrp.AppBusiness{},
				expResp: toAppBusinessPtr(sd.businesses[2]),
			},
			{
				name:    "appointment",
				url:     fmt.Sprintf("/v1/appointments/%s", sd.appointments[1].ID),
				resp:    &appointmentgrp.AppAppointment{},
				expResp: toAppAppointmentPtr(sd.appointments[1]),
			},
			{
				name:    "general_agenda",
				url:     fmt.Sprintf("/v1/agendas/general/%s", sd.generalAgendas[0].ID),
				resp:    &agendagrp.AppGeneralAgenda{},
				expResp: toAppGeneralAgendaPtr(sd.generalAgendas[0]),
			},
			{
				name:    "daily_agenda",
				url:     fmt.Sprintf("/v1/agendas/daily/%s", sd.dailyAgendas[0].ID),
				resp:    &agendagrp.AppDailyAgenda{},
				expResp: toAppDailyAgendaPtr(sd.dailyAgendas[0]),
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
					Enabled: false,
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
	sch := sd.generalAgendas[0].OpensAt.Add(1 * time.Hour).UTC().Format(time.RFC3339)

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
					ScheduledOn: sch,
				},
				resp: &appointmentgrp.AppAppointment{},
				expResp: &appointmentgrp.AppAppointment{
					BusinessID:  sd.businesses[0].ID.String(),
					UserID:      sd.users[0].ID.String(),
					Status:      appointment.StatusScheduled.Status(),
					ScheduledOn: sch,
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

func (wt *WebTests) createGeneralAgenda200(sd seedData) func(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now()
	return func(t *testing.T) {
		table := []struct {
			name    string
			url     string
			input   any
			resp    any
			expResp any
		}{
			{
				name: "general_agenda",
				url:  "/v1/agendas/general",
				input: &agendagrp.AppNewGeneralAgenda{
					BusinessID:  sd.businesses[1].ID.String(),
					OpensAt:     time.Date(now.Year(), now.Month(), now.Day(), 10, 12, 0, 0, loc).Format(time.RFC3339),
					ClosedAt:    time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc).Format(time.RFC3339),
					Interval:    2 * 60 * 60, // Every 2 hours
					WorkingDays: []int{1, 5},
				},
				resp: &agendagrp.AppGeneralAgenda{},
				expResp: &agendagrp.AppGeneralAgenda{
					BusinessID:  sd.businesses[1].ID.String(),
					OpensAt:     time.Date(now.Year(), now.Month(), now.Day(), 10, 12, 0, 0, loc).Format(time.RFC3339),
					ClosedAt:    time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc).Format(time.RFC3339),
					Interval:    2 * 60 * 60, // Every 2 hours
					WorkingDays: []int{1, 5},
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
				t.Errorf("%s: Should be able to unmarshal the respones: %s", tt.name, err)
			}

			gotResp, exists := tt.resp.(*agendagrp.AppGeneralAgenda)
			if !exists {
				t.Fatalf("error occurred")
			}

			expResp := tt.expResp.(*agendagrp.AppGeneralAgenda)
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

func (wt *WebTests) createDailyAgenda200(sd seedData) func(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now()
	scheduledDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 2)

	return func(t *testing.T) {
		table := []struct {
			name    string
			url     string
			input   any
			resp    any
			expResp any
		}{
			{
				name: "daily_agenda",
				url:  "/v1/agendas/daily",
				input: &agendagrp.AppNewDailyAgenda{
					BusinessID:   sd.businesses[2].ID.String(),
					OpensAt:      time.Date(now.Year(), now.Month(), now.Day(), 10, 12, 0, 0, loc).Format(time.RFC3339),
					ClosedAt:     time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc).Format(time.RFC3339),
					Interval:     2 * 60 * 60, // Every 2 hours
					Date:         scheduledDate.Format(time.RFC3339),
					Availability: true,
				},
				resp: &agendagrp.AppDailyAgenda{},
				expResp: &agendagrp.AppDailyAgenda{
					BusinessID:   sd.businesses[2].ID.String(),
					OpensAt:      time.Date(now.Year(), now.Month(), now.Day(), 10, 12, 0, 0, loc).Format(time.RFC3339),
					ClosedAt:     time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc).Format(time.RFC3339),
					Interval:     2 * 60 * 60, // Every 2 hours
					Date:         scheduledDate.Format(time.DateOnly),
					Availability: true,
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
				t.Errorf("%s: Should be able to unmarshal the respones: %s", tt.name, err)
			}

			gotResp, exists := tt.resp.(*agendagrp.AppGeneralAgenda)
			if !exists {
				t.Fatalf("error occurred")
			}

			expResp := tt.expResp.(*agendagrp.AppGeneralAgenda)
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
