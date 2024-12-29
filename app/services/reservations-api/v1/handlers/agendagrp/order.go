package agendagrp

import (
	"github.com/ameghdadian/service/business/core/agenda"
)

var generalAgendaOrderByFields = map[string]string{
	"id":          agenda.OrderByID,
	"business_id": agenda.OrderByBusinessID,
}

var dailyAgendaOrderByFields = map[string]string{
	"id":          agenda.OrderByID,
	"business_id": agenda.OrderByBusinessID,
}
