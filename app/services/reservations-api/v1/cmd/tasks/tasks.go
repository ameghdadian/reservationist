package tasks

import (
	"github.com/ameghdadian/service/app/services/reservations-api/v1/handlers/appointmentgrp"
	v1 "github.com/ameghdadian/service/business/web/v1"
)

func Handlers() add {
	return add{}
}

type add struct{}

func (add) Add(cfg v1.TaskMuxConfig) {
	appointmentgrp.RegisterTaskHandlers(appointmentgrp.TaskConfig{
		DB:  cfg.DB,
		Log: cfg.Log,
		Mux: cfg.Mux,
	})
}
