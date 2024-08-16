package testgrp

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"

	"github.com/ameghdadian/service/business/web/v1/response"
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

func (h *Handlers) TestGrpHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if rand.Float64() > 0.5 {
		h.log.Error(ctx, "bad happend")
		return response.NewError(errors.New("TRUSTED ERROR"), http.StatusBadRequest)
	}

	h.log.Info(ctx, "Succesful", "build info:", h.build)
	return web.Respond(ctx, w, "All good", http.StatusOK)
}
