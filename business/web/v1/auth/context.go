package auth

import (
	"context"

	"github.com/google/uuid"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// claimKey is used to store/retrieve a Claims value from a context.Context
const claimKey ctxKey = 1

// usrKey is used to store/retrive a user value from a context.Context
const usrKey ctxKey = 2

func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

func GetClaims(ctx context.Context) Claims {
	v, ok := ctx.Value(claimKey).(Claims)
	if !ok {
		return Claims{}
	}

	return v
}

func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, usrKey, userID)
}

func GetUserID(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(usrKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}
	}

	return v
}
