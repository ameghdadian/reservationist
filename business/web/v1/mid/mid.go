package mid

import (
	"context"
	"errors"

	"github.com/ameghdadian/service/business/core/agenda"
	"github.com/ameghdadian/service/business/core/appointment"
	"github.com/ameghdadian/service/business/core/business"
	"github.com/ameghdadian/service/business/core/user"
)

type ctxKey int

const (
	userKey ctxKey = iota
	businesesKey
	appointmentKey
	generalAgendaKey
	dailyAgendaKey
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

func setGeneralAgenda(ctx context.Context, agd agenda.GeneralAgenda) context.Context {
	return context.WithValue(ctx, generalAgendaKey, agd)
}

func GetGeneralAgenda(ctx context.Context) (agenda.GeneralAgenda, error) {
	v, ok := ctx.Value(generalAgendaKey).(agenda.GeneralAgenda)
	if !ok {
		return agenda.GeneralAgenda{}, errors.New("agenda not found in context")
	}

	return v, nil
}

func setDailyAgenda(ctx context.Context, agd agenda.DailyAgenda) context.Context {
	return context.WithValue(ctx, dailyAgendaKey, agd)
}

func GetDailyAgenda(ctx context.Context) (agenda.DailyAgenda, error) {
	v, ok := ctx.Value(dailyAgendaKey).(agenda.DailyAgenda)
	if !ok {
		return agenda.DailyAgenda{}, errors.New("agenda not found in context")
	}

	return v, nil
}
