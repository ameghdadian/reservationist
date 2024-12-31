package mux

import (
	"context"

	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/mid"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

type APIMuxConfig struct {
	Build         string
	Log           *logger.Logger
	Auth          *auth.Auth
	DB            *sqlx.DB
	Tracer        trace.Tracer
	TaskClient    *asynq.Client
	TaskInspector *asynq.Inspector
}

type RouterAdder interface {
	Add(app *web.App, cfg APIMuxConfig)
}

func APIMux(cfg APIMuxConfig, routeAdder RouterAdder) *web.App {
	logger := func(ctx context.Context, msg string, args ...any) {
		cfg.Log.Info(ctx, msg, args...)
	}
	app := web.NewApp(
		logger,
		cfg.Tracer,
		mid.Otel(cfg.Tracer),
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
