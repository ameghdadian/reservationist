package all

import (
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/checkgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/testgrp"
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/usergrp"
	v1 "github.com/ameghdadian/service/business/web/v1"
	"github.com/ameghdadian/service/foundation/web"
)

func Routes() add {
	return add{}
}

type add struct{}

func (add) Add(app *web.App, cfg v1.APIMuxConfig) {

	testgrp.Routes(app, testgrp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		Auth:  cfg.Auth,
	})

	checkgrp.Routes(app, checkgrp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	usergrp.Routes(app, usergrp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
		Auth:  cfg.Auth,
	})
}
