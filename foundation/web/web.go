package web

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	mux    *http.ServeMux
	otmux  http.Handler
	mw     []MidFunc
}

func NewApp(log Logger, tracer trace.Tracer, mw ...MidFunc) *App {
	mux := http.NewServeMux()

	return &App{
		log:    log,
		tracer: tracer,
		mux:    mux,
		otmux:  otelhttp.NewHandler(mux, "http request"),
		mw:     mw,
	}
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This was set up in the NewApp function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.otmux.ServeHTTP(w, r)
}

// HandleNoMiddleware sets a handler function for a given HTTP method and path
// pair to the application server mux. Does not include the application
// middleware or OTEL tracing.
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

	a.mux.HandleFunc(pattern, h)
}
