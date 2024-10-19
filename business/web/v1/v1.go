package v1

import (
	"os"

	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/mid"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

type APIMuxConfig struct {
	Build         string
	Shutdown      chan os.Signal
	Log           *logger.Logger
	Auth          *auth.Auth
	DB            *sqlx.DB
	TaskClient    *asynq.Client
	TaskInspector *asynq.Inspector
}

type RouterAdder interface {
	Add(app *web.App, cfg APIMuxConfig)
}

func APIMux(cfg APIMuxConfig, routeAdder RouterAdder) *web.App {
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	routeAdder.Add(app, cfg)

	return app
}

type TaskMuxConfig struct {
	DB  *sqlx.DB
	Log *logger.Logger
	Mux *asynq.ServeMux
}

type TaskRouter interface {
	Add(cfg TaskMuxConfig)
}

func TaskMux(cfg TaskMuxConfig, taskAdder TaskRouter) {
	taskAdder.Add(cfg)
}
