package authgrp

import (
	"encoding/json"

	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/google/uuid"
)

type token struct {
	Token string `json:"token"`
}

func (t token) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	return data, "application/json", err
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

func (a authenticateResp) Encode() ([]byte, string, error) {
	data, err := json.Marshal(a)
	return data, "application/json", err
}
