package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Encoder defines behaviour that can encode a data model and provide the
// content type for that encoding.
type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

// Handler represents a function that handles a http request within our framework.
// Both httptreemux and go router provided following handler signature
// func (w http.ResponseWriter, r *http.Request)
// What we need:
// func (ctx context.Context, r *http.Requst) error
// So we need a customized version of handler.
type HandlerFunc func(ctx context.Context, r *http.Request) Encoder

type Logger func(ctx context.Context, msg string, args ...any)
type App struct {
	log    Logger
	tracer trace.Tracer
	*http.ServeMux
	mw []MidFunc
}

func NewApp(log Logger, tracer trace.Tracer, mw ...MidFunc) *App {
	return &App{
		log:      log,
		tracer:   tracer,
		ServeMux: http.NewServeMux(),
		mw:       mw,
	}
}

func (a *App) HandleNoMiddleware(method string, group string, path string, handler HandlerFunc) {
	a.handle(method, group, path, handler)
}

func (a *App) Handle(method string, group string, path string, handler HandlerFunc, mw ...MidFunc) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	a.handle(method, group, path, handler)
}

func (a *App) handle(method string, group string, path string, handler HandlerFunc) {

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := setTracer(r.Context(), a.tracer)
		ctx = setWriter(ctx, w)

		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

		v := Values{
			TraceID: uuid.NewString(),
			Now:     time.Now().UTC(),
		}
		ctx = SetValues(ctx, &v)

		resp := handler(ctx, r)
		if err := Respond(ctx, w, r, resp); err != nil {
			a.log(ctx, "web-respond", "ERROR", err)
			return
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	pattern := fmt.Sprintf("%s %s", method, finalPath)

	a.ServeMux.HandleFunc(pattern, h)
}
