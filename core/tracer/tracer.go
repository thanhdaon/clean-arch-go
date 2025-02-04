package tracer

import (
	"context"
	"crypto/tls"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func SetupTracer() func() {
	ctx := context.Background()

	exporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	exporterApiToken := os.Getenv("OTEL_EXPORTER_OTLP_APITOKEN")

	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(exporterEndpoint),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization":   "Bearer " + exporterApiToken,
			"X-AXIOM-DATASET": "task-traces",
		}),
		otlptracehttp.WithTLSClientConfig(&tls.Config{}),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		logrus.Fatalf("Failed to create OTLP trace exporter: %v", err)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("tasks"),
		semconv.ServiceVersionKey.String("0.1.0"),
		semconv.DeploymentEnvironmentKey.String("staging"),
	)

	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logrus.Fatalf("Failed to shut down tracer provider: %v", err)
		}
	}
}
