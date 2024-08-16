package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ameghdadian/service/business/web/debug"
	v1 "github.com/ameghdadian/service/business/web/v1"
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/foundation/keystore"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/ardanlabs/conf/v3"
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

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "SALES-API", traceIDFunc, events)

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
	// Configuration

	cfg := struct {
		conf.Version
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

	const prefix = "SALES"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

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
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", ":8000")

		if err := http.ListenAndServe(":8000", debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", ":8000", "msg", err)
		}
	}()

	// ------------------------------------------------------------------------------
	// Start API Service

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := v1.APIMuxConfig{
		Build:    build,
		Shutdown: shutdown,
		Log:      log,
		Auth:     auth,
	}

	apiMux := v1.APIMux(cfgMux, routeAdder)

	srv := http.Server{
		Addr:    ":9000",
		Handler: apiMux,
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

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
		log.Info(ctx, "Received shutdown signal, exitting")
	}

	return nil
}
