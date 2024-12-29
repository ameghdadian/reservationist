package checkgrp

import (
	"context"
	"net/http"
	"os"
	"time"

	db "github.com/ameghdadian/service/business/data/dbsql/pgx"
	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

type handlers struct {
	build string
	log   *logger.Logger
	db    *sqlx.DB
}

func newApp(build string, log *logger.Logger, db *sqlx.DB) *handlers {
	return &handlers{
		build: build,
		log:   log,
		db:    db,
	}
}

// Readiness checks application's readiness to accept connections.
// Failing Readiness will cause no traffic to be routed to the application.
func (h *handlers) readiness(ctx context.Context, r *http.Request) web.Encoder {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := db.StatusCheck(ctx, h.db); err != nil {
		h.log.Info(ctx, "readiness failure", "ERROR", err)
		return errs.New(errs.Internal, err)
	}

	return nil
}

// Liveness checks whether the application is live.
// Failing Liveness will cause the pod to be restarted.
func (h *handlers) liveness(ctx context.Context, r *http.Request) web.Encoder {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	info := Info{
		Status:     "up",
		Build:      h.build,
		Host:       host,
		Name:       os.Getenv("KUBERNETES_NAME"),
		PodIP:      os.Getenv("KUBERNETES_POD_IP"),
		Node:       os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace:  os.Getenv("KUBERNETES_NAMESPACE"),
		GOMAXPROCS: os.Getenv("GOMAXPROCS"),
	}

	return info
}
