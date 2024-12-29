package appointmentgrp

import (
	"net/http"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/agenda/stores/agendadb"
	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/appointment/stores/appointmentdb"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/business/stores/businessdb"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/core/user/stores/userdb"
	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/mid"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Build         string
	Log           *logger.Logger
	DB            *sqlx.DB
	Auth          *auth.Auth
	TaskClient    *asynq.Client
	TaskInspector *asynq.Inspector
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	aptTask := appointment.NewTask(cfg.TaskClient, cfg.TaskInspector)

	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	bsnCore := business.NewCore(cfg.Log, usrCore, businessdb.NewStore(cfg.Log, cfg.DB))
	aptCore := appointment.NewCore(cfg.Log, usrCore, bsnCore, appointmentdb.NewStore(cfg.Log, cfg.DB), aptTask)
	agdCore := agenda.NewCore(cfg.Log, agendadb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	ruleAdminOnly := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAuthorizeAppointment := mid.AuthorizeAppointment(cfg.Log, cfg.Auth, aptCore)
	tran := mid.ExecuteInTransaction(cfg.Log, db.NewBeginner(cfg.DB))

	hdl := newApp(aptCore, agdCore)
	app.Handle(http.MethodGet, version, "/appointments", hdl.query, authen, ruleAdminOnly)
	app.Handle(http.MethodGet, version, "/appointments/{appointment_id}", hdl.queryByID, authen, ruleAuthorizeAppointment)
	app.Handle(http.MethodPost, version, "/appointments", hdl.create, authen, tran)
	app.Handle(http.MethodPut, version, "/appointments/{appointment_id}", hdl.update, authen, tran, ruleAuthorizeAppointment)
	app.Handle(http.MethodDelete, version, "/appointments/{appointment_id}", hdl.delete, authen, tran, ruleAuthorizeAppointment)
}
