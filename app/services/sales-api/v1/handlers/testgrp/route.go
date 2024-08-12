package testgrp

import (
	"net/http"

	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
)

type Config struct {
	Build string
	Log   *logger.Logger
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	hdl := New(cfg.Build, cfg.Log)
	app.Handle(http.MethodGet, version, "", hdl.TestGrpHandler)
}
