package usergrp

import (
	"github.com/ameghdadian/service/business/core/user"
)

var orderByFields = map[string]string{
	"user_id":      user.OrderByID,
	"email":        user.OrderByEmail,
	"name":         user.OrderByName,
	"phone_number": user.OrderByPhoneNumber,
	"roles":        user.OrderByRoles,
	"enabled":      user.OrderByEnabled,
}
