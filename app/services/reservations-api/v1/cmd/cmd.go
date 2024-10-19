package cmd

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ameghdadian/service/business/web/debug"
	v1 "github.com/ameghdadian/service/business/web/v1"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/foundation/keystore"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/ardanlabs/conf/v3"
	"github.com/hibiken/asynq"
)

func Main(build string, routeAdder v1.RouterAdder) error {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "************ SEND ALERT ************")
		},
	}

	traceIDFunc := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "RESERVATIONS-API", traceIDFunc, events)

	// -----------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log, build, routeAdder); err != nil {
		log.Error(ctx, "startup", "msg", err)
		return err
	}

	return nil
}

func run(ctx context.Context, log *logger.Logger, build string, routeAdder v1.RouterAdder) error {

	// ------------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0), "build", build)

	// ------------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s,mask"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:database-service.reservations-system.svc.cluster.local"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Redis struct {
			Addr string `conf:"default:0.0.0.0:6379"`
		}
		Auth struct {
			KeysFolder string `conf:"default:zarf/keys/"`
			ActiveKID  string `conf:"963df661-d92e-4991-b519-77d838a21705"`
			Issuer     string `conf:"default:service project"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "AMIR. ME.",
		},
	}

	const prefix = "RESERVATIONS"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// ------------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	expvar.NewString("build").Set(build)

	// ------------------------------------------------------------------------------
	// Initialize database support

	log.Info(ctx, "startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := db.Open(db.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping database support", "host", cfg.DB.Host)
	}()

	// ------------------------------------------------------------------------------
	// Initialize async task scheduler support

	opt := asynq.RedisClientOpt{Addr: cfg.Redis.Addr}
	taskClient := asynq.NewClient(opt)
	defer taskClient.Close()

	taskInspector := asynq.NewInspector(opt)
	defer taskInspector.Close()

	// ------------------------------------------------------------------------------
	// Initialize authentication support

	log.Info(ctx, "startup", "status", "initializing authentication support")

	// Using keystore to store JWT private key
	ks, err := keystore.NewFS(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	authCfg := auth.Config{
		Log:       log,
		KeyLookup: ks,
	}

	auth, err := auth.New(authCfg)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	// ------------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// ------------------------------------------------------------------------------
	// Start API Service

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := v1.APIMuxConfig{
		Build:         build,
		Shutdown:      shutdown,
		Log:           log,
		Auth:          auth,
		DB:            db,
		TaskClient:    taskClient,
		TaskInspector: taskInspector,
	}

	apiMux := v1.APIMux(cfgMux, routeAdder)

	srv := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info(ctx, "Server is running on", "addr", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// ------------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		log.Info(ctx, "Error while running server", "err", err)
	case <-shutdown:

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
		log.Info(ctx, "Received shutdown signal, exitting")
	}

	return nil
}

func InitTaskWorkers(build string, taskRouter v1.TaskRouter) error {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "************ SEND ALERT ************")
		},
	}

	traceIDFunc := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "RESERVATIONS-API", traceIDFunc, events)

	// -----------------------------------------------------------------------

	ctx := context.Background()

	if err := initWorkers(ctx, log, build, taskRouter); err != nil {
		log.Error(ctx, "startup", "msg", err)
		return err
	}

	return nil
}

func initWorkers(ctx context.Context, log *logger.Logger, build string, taskRouter v1.TaskRouter) error {

	// ------------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0), "build", build)

	// ------------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Worker struct {
			ShutdownTimeout time.Duration `conf:"default:20s"`
			NumOfWorkers    int           `conf:"default:10"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:database-service.reservations-system.svc.cluster.local"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Redis struct {
			Addr string `conf:"default:0.0.0.0:6379"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "AMIR. ME.",
		},
	}

	const prefix = "TASKS"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// ------------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	expvar.NewString("build").Set(build)

	// ------------------------------------------------------------------------------
	// Initialize database support

	log.Info(ctx, "startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := db.Open(db.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping database support", "host", cfg.DB.Host)
	}()

	// ------------------------------------------------------------------------------
	// Initialize task worker server

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		asynq.Config{
			BaseContext:     func() context.Context { return ctx },
			Concurrency:     cfg.Worker.NumOfWorkers,
			ShutdownTimeout: cfg.Worker.ShutdownTimeout,
		},
	)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	mux := asynq.NewServeMux()
	cfgMux := v1.TaskMuxConfig{
		DB:  db,
		Log: log,
		Mux: mux,
	}
	v1.TaskMux(cfgMux, taskRouter)

	serverErrors := make(chan error, 1)
	go func() {
		log.Info(ctx, "Number of workers:", "count", cfg.Worker.NumOfWorkers)
		serverErrors <- srv.Run(mux)
	}()

	// ------------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		log.Info(ctx, "Error while running server", "err", err)

	case <-shutdown:
		log.Info(ctx, "Received shutdown signal, exitting")
		srv.Shutdown()
	}

	return nil
}
