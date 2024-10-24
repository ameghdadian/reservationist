package agendagrp

import (
	"net/http"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/agenda/stores/agendadb"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/business/stores/businessdb"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/core/user/stores/userdb"
	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/mid"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Build string
	Log   *logger.Logger
	DB    *sqlx.DB
	Auth  *auth.Auth
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	bsnCore := business.NewCore(cfg.Log, usrCore, businessdb.NewStore(cfg.Log, cfg.DB))
	agdCore := agenda.NewCore(cfg.Log, bsnCore, agendadb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	ruleAdminOnly := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAuthorizedGenAgenda := mid.AuthorizeGeneralAgenda(cfg.Log, cfg.Auth, agdCore, bsnCore)
	ruleAuthorizedDaiAgenda := mid.AuthorizeDailyAgenda(cfg.Log, cfg.Auth, agdCore, bsnCore)
	tran := mid.ExecuteInTransaction(cfg.Log, db.NewBeginner(cfg.DB))

	hdl := New(agdCore)
	// General Agenda Handlers
	app.Handle(http.MethodPost, version, "/agendas/general", hdl.CreateGeneralAgenda, authen, tran)
	app.Handle(http.MethodPut, version, "/agendas/general/{agenda_id}", hdl.UpdateGeneralAgenda, authen, tran, ruleAuthorizedGenAgenda)
	app.Handle(http.MethodDelete, version, "/agendas/general/{agenda_id}", hdl.DeleteGeneralAgenda, authen, tran, ruleAuthorizedGenAgenda)
	app.Handle(http.MethodGet, version, "/agendas/general", hdl.QueryGeneralAgenda, authen, ruleAdminOnly)
	app.Handle(http.MethodGet, version, "/agendas/general/{agenda_id}", hdl.QueryGeneralAgendaByID, authen)
	// Daily Agenda Handlers
	app.Handle(http.MethodPost, version, "/agendas/daily", hdl.CreateDailyAgenda, authen, tran)
	app.Handle(http.MethodPut, version, "/agendas/daily/{agenda_id}", hdl.UpdateDailyAgenda, authen, tran, ruleAuthorizedDaiAgenda)
	app.Handle(http.MethodDelete, version, "/agendas/daily/{agenda_id}", hdl.DeleteDailyAgenda, authen, tran, ruleAuthorizedDaiAgenda)
	app.Handle(http.MethodGet, version, "/agendas/daily", hdl.QueryDailyAgenda, authen)
	app.Handle(http.MethodGet, version, "/agendas/daily/{agenda_id}", hdl.QueryDailyAgendaByID, authen)
}
