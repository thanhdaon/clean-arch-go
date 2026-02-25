package ports

import (
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"context"
	"net/http"

	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/trace"
)

func MapErrorToStatus(err error) int {
	switch {
	case errors.Is(errkind.Authorization, err):
		return http.StatusUnauthorized
	case errors.Is(errkind.Permission, err):
		return http.StatusForbidden
	case errors.Is(errkind.NotExist, err):
		return http.StatusNotFound
	case errors.Is(errkind.Exist, err):
		return http.StatusConflict
	case errors.Is(errkind.Connection, err):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func ResponseError(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	status := MapErrorToStatus(err)
	render.Respond(w, r, map[string]any{
		"error":    err.Error(),
		"status":   status,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}

func internalError(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	ResponseError(ctx, errors.E(errkind.Internal, err), w, r)
}

func unauthorised(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	ResponseError(ctx, err, w, r)
}

func badRequest(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	render.Respond(w, r, map[string]any{
		"error":    err.Error(),
		"status":   http.StatusBadRequest,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}

func responseSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	render.Respond(w, r, map[string]any{
		"status":   http.StatusOK,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}
