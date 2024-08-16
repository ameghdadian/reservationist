package testgrp

import (
	"net/http"

	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/mid"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
)

type Config struct {
	Build string
	Log   *logger.Logger
	Auth  *auth.Auth
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Auth)
	ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)

	hdl := New(cfg.Build, cfg.Log)
	app.Handle(http.MethodGet, version, "", hdl.TestGrpHandler, authen, ruleAdmin)
}
