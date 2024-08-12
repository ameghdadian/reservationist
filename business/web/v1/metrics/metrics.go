package metrics

import (
	"context"
	"expvar"
	"runtime"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. The expvar package is already based on singleton
// for different metrics that are registered with the package so there
// isn't much choice here.
var m *metrics

type metrics struct {
	goroutines *expvar.Int
	requests   *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

func init() {
	m = &metrics{
		goroutines: expvar.NewInt("goroutines"),
		requests:   expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}
}

type ctxKey int

const key ctxKey = 1

// Set sets the metrics data into the contxt.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

func AddGoroutines(ctx context.Context) int64 {
	if v, ok := ctx.Value(ctx).(*metrics); ok {
		if v.requests.Value()%100 == 0 {
			g := int64(runtime.NumGoroutine())
			v.goroutines.Set(g)
			return g
		}
	}

	return 0
}

func AddRequests(ctx context.Context) int64 {
	if v, ok := ctx.Value(ctx).(*metrics); ok {
		v.requests.Add(1)
		return v.requests.Value()
	}

	return 0
}

func AddErrors(ctx context.Context) int64 {
	if v, ok := ctx.Value(ctx).(*metrics); ok {
		v.errors.Add(1)
		return v.errors.Value()
	}

	return 0
}

func AddPanics(ctx context.Context) int64 {
	if v, ok := ctx.Value(ctx).(*metrics); ok {
		v.panics.Add(1)
		return v.panics.Value()
	}

	return 0
}
