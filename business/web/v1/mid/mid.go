package mid

import (
	"context"
	"errors"

	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
)

type ctxKey int

const (
	userKey ctxKey = iota
	businesesKey
	appointmentKey
)

func setUser(ctx context.Context, usr user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

func GetUser(ctx context.Context) (user.User, error) {
	v, ok := ctx.Value(userKey).(user.User)
	if !ok {
		return user.User{}, errors.New("user not found in context")
	}

	return v, nil
}

func setBusiness(ctx context.Context, bsn business.Business) context.Context {
	return context.WithValue(ctx, businesesKey, bsn)
}

func GetBusiness(ctx context.Context) (business.Business, error) {
	v, ok := ctx.Value(businesesKey).(business.Business)
	if !ok {
		return business.Business{}, errors.New("business not found in context")
	}

	return v, nil
}

func setAppointment(ctx context.Context, apt appointment.Appointment) context.Context {
	return context.WithValue(ctx, appointmentKey, apt)
}

func GetAppointment(ctx context.Context) (appointment.Appointment, error) {
	v, ok := ctx.Value(appointmentKey).(appointment.Appointment)
	if !ok {
		return appointment.Appointment{}, errors.New("appointment not found in context")
	}

	return v, nil
}
