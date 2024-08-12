package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// Both httptreemux and go router provided following handler signature
// func (w http.ResponseWriter, r *http.Request)
// What we need:
// func (ctx context.Context, w http.ResponseWriter, r *http.Requst) error
// So we need a customized version of handler.

type App struct {
	*http.ServeMux
	shutdown chan os.Signal
	mw       []Middleware
}

func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ServeMux: http.NewServeMux(),
		shutdown: shutdown,
		mw:       mw,
	}
}

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) HandleNoMiddleware(method string, group string, path string, handler Handler) {
	a.handle(method, group, path, handler)
}

func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	a.handle(method, group, path, handler)
}

func (a *App) handle(method string, group string, path string, handler Handler) {
	h := func(w http.ResponseWriter, r *http.Request) {

		v := Values{
			TraceID: uuid.NewString(),
			Now:     time.Now().UTC(),
		}
		ctx := SetValues(r.Context(), &v)

		if err := handler(ctx, w, r); err != nil {
			if validateShutdown(err) {
				a.SignalShutdown()
				return
			}
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	pattern := fmt.Sprintf("%s %s", method, finalPath)

	a.ServeMux.HandleFunc(pattern, h)
}

func validateShutdown(err error) bool {
	switch {
	// You get broken pipe error when writing on the connection that's
	// previously closed.
	case errors.Is(err, syscall.EPIPE):
		return false
	// You get connection reset by peer error when reading from connection
	// closed by peer.
	case errors.Is(err, syscall.ECONNRESET):
		return false
	}
	return true
}
