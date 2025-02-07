package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ameghdadian/service/foundation/errs"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/web"
)

func Logger(log *logger.Logger) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, r *http.Request) web.Encoder {
			now := time.Now()

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Info(ctx, "request started", "method", r.Method, "path", path,
				"remoteaddr", r.RemoteAddr)

			resp := next(ctx, r)
			err := isError(resp)

			var statusCode = errs.OK
			if err != nil {
				statusCode = errs.Internal

				var v *errs.Error
				if errors.As(err, &v) {
					statusCode = v.Code
				}
			}

			log.Info(ctx, "request completed", "method", r.Method, "path", path,
				"remoteaddr", r.RemoteAddr, "statuscode", statusCode, "since", time.Since(now))

			return resp
		}

		return h
	}

	return m
}
