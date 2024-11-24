package authgrp

import (
	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/google/uuid"
)

type token struct {
	Token string `json:"token"`
}

func toToken(v string) token {
	return token{
		Token: v,
	}
}

type authenticateResp struct {
	UserID uuid.UUID   `json:"user_id"`
	Claims auth.Claims `json:"claims"`
}
