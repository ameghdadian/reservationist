package tasks

import (
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/appointmentgrp"
	"github.com/ameghdadian/service/business/web/v1/mux"
)

func Handlers() add {
	return add{}
}

type add struct{}

func (add) Add(cfg mux.TaskMuxConfig) {
	appointmentgrp.RegisterTaskHandlers(appointmentgrp.TaskConfig{
		DB:  cfg.DB,
		Log: cfg.Log,
		Mux: cfg.Mux,
	})
}
