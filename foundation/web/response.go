package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

type NoResponse struct{}

func NewNoResponse() NoResponse {
	return NoResponse{}
}

func (NoResponse) Encode() ([]byte, string, error) {
	return nil, "", nil
}

// =======================================================================
type httpStatus interface {
	HTTPStatus() int
}

// Respond sends a response to the client.
func Respond(ctx context.Context, w http.ResponseWriter, r *http.Request, dataModel Encoder) error {
	if _, ok := dataModel.(NoResponse); ok {
		return nil
	}

	// If context is canceled, it means client is no longer waiting for a response.
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}

	var statusCode int
	switch r.Method {
	case http.MethodPost:
		statusCode = http.StatusCreated
	default:
		statusCode = http.StatusOK
	}

	switch v := dataModel.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()

	case error:
		statusCode = http.StatusInternalServerError

	default:
		if dataModel == nil {
			statusCode = http.StatusNoContent
		}
	}

	_, span := addSpan(ctx, "web.send.response", attribute.Int("status", statusCode))
	defer span.End()

	SetStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, contentType, err := dataModel.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}
