package businessgrp

import (
	"github.com/ameghdadian/service/business/core/business"
)

var orderByFields = map[string]string{
	"business_id": business.OrderByID,
	"owner_id":    business.OrderByOwnerID,
	"name":        business.OrderByName,
	"description": business.OrderByDesc,
}
