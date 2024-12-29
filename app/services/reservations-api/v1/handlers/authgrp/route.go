package authgrp

import (
	"net/http"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/core/user/stores/userdb"
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

	bearer := mid.Bearer(cfg.Auth)
	basic := mid.Basic(cfg.Auth, usrCore)

	hdl := newApp(usrCore, cfg.Auth)
	app.Handle(http.MethodGet, version, "/auth/token/{kid}", hdl.token, basic)
	app.Handle(http.MethodGet, version, "/auth/authenticate", hdl.authenticate, bearer)
}
