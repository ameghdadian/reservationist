package businessgrp

import (
	"net/http"

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
	Auth  *auth.Auth
	DB    *sqlx.DB
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	bsnCore := business.NewCore(cfg.Log, usrCore, businessdb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	ruleAuthorizeBusiness := mid.AuthorizeBusiness(cfg.Log, cfg.Auth, bsnCore)
	tran := mid.ExecuteInTransaction(cfg.Log, db.NewBeginner(cfg.DB))

	hdl := newApp(bsnCore, usrCore)
	app.Handle(http.MethodGet, version, "/businesses", hdl.query, authen)
	app.Handle(http.MethodGet, version, "/businesses/{business_id}", hdl.queryByID, authen)
	app.Handle(http.MethodPost, version, "/businesses", hdl.create, authen, tran)
	app.Handle(http.MethodPut, version, "/businesses/{business_id}", hdl.update, authen, tran, ruleAuthorizeBusiness)
	app.Handle(http.MethodDelete, version, "/businesses/{business_id}", hdl.delete, authen, tran, ruleAuthorizeBusiness)
}
