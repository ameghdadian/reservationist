package checkgrp

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
)

type Handlers struct {
	build string
	log   *logger.Logger
}

func New(build string, log *logger.Logger) *Handlers {
	return &Handlers{
		build: build,
		log:   log,
	}
}

// Readiness checks application's readiness to accept connections.
// Failing Readiness will cause no traffic to be routed to the application.
func (h *Handlers) Readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	status := "ok"
	statusCode := http.StatusOK

	data := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	h.log.Info(ctx, "readiness", "status", status)

	return web.Respond(ctx, w, data, statusCode)
}

// Liveness checks whether the application is live.
// Failing Liveness will cause the pod to be restarted.
func (h *Handlers) Liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	data := struct {
		Status     string `json:"status,omitempty"`
		Build      string `json:"build,omitempty"`
		Host       string `json:"host,omitempty"`
		Name       string `json:"name,omitempty"`
		PodIP      string `json:"podIP,omitempty"`
		Node       string `json:"node,omitempty"`
		Namespace  string `json:"namespace,omitempty"`
		GOMAXPROCS string `json:"GOMAXPROCS,omitempty"`
	}{
		Status:     "up",
		Build:      h.build,
		Host:       host,
		Name:       os.Getenv("KUBERNETES_NAME"),
		PodIP:      os.Getenv("KUBERNETES_POD_IP"),
		Node:       os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace:  os.Getenv("KUBERNETES_NAMESPACE"),
		GOMAXPROCS: os.Getenv("GOMAXPROCS"),
	}

	h.log.Info(ctx, "liveness", "status", "OK")

	return web.Respond(ctx, w, data, http.StatusOK)
}
