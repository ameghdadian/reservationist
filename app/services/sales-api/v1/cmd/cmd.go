package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ameghdadian/service/business/web/debug"
	v1 "github.com/ameghdadian/service/business/web/v1"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
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
