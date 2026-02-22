package tracer

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func SetupTracer() func() {
	ctx := context.Background()

	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		logrus.Fatalln("Missing JAEGER_ENDPOINT env")
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(jaegerEndpoint),
	)
	if err != nil {
		logrus.Fatalf("Failed to create OTLP HTTP exporter: %v", err)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("tasks"),
	)

	tracerProvider := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
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
