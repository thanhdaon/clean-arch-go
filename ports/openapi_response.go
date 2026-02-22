package ports

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/trace"
)

func internalError(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	responseError(ctx, err, w, r, http.StatusInternalServerError)
}

func unauthorised(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	responseError(ctx, err, w, r, http.StatusUnauthorized)
}

func badRequest(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	responseError(ctx, err, w, r, http.StatusBadRequest)
}

func responseError(ctx context.Context, err error, w http.ResponseWriter, r *http.Request, status int) {
	render.Respond(w, r, map[string]any{
		"error":    err.Error(),
		"status":   status,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}

func responseSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	render.Respond(w, r, map[string]any{
		"status":   http.StatusOK,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}
