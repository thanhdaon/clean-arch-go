package adapters

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewHttpClient() *http.Client {
	spanNameFormatter := otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
		return fmt.Sprintf("HTTP %s %s %s", r.Method, r.URL.String(), operation)
	})

	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport, spanNameFormatter),
	}
}
