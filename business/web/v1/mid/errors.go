package mid

import (
	"context"
	"net/http"

	"github.com/ameghdadian/service/business/web/v1/auth"
	"github.com/ameghdadian/service/business/web/v1/response"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/ameghdadian/service/foundation/validate"
	"github.com/ameghdadian/service/foundation/web"
)

func Errors(log *logger.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			if err := handler(ctx, w, r); err != nil {
				log.Error(ctx, "message", "msg", err)

				var er response.ErrorDocument
				var status int

				switch {
				// Trusted errors
				case response.IsError(err):
					reqErr := response.GetError(err)
					er = response.ErrorDocument{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				// Field error
				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = response.ErrorDocument{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				// Auth error
				case auth.IsAuthError(err):
					er = response.ErrorDocument{
						Error: http.StatusText(http.StatusUnauthorized),
					}
					status = http.StatusUnauthorized

				// Untrusted errors
				default:
					er = response.ErrorDocument{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shutdown the service.
				if web.IsShutdown(err) {
					return err
				}
			}
			return nil
		}

		return h
	}

	return m
}
