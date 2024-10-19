package appointmentgrp

import (
	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/business/core/user/stores/userdb"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

type TaskConfig struct {
	DB  *sqlx.DB
	Log *logger.Logger
	Mux *asynq.ServeMux
}

func RegisterTaskHandlers(cfg TaskConfig) {
	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))

	th := appointment.NewTaskHandlers(cfg.Log, usrCore)

	cfg.Mux.HandleFunc(appointment.TypeSendSMS, th.HandleSendSMS)
}
