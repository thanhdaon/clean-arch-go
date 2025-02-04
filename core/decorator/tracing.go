package decorator

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type commandTracingDecorator[C any] struct {
	base CommandHandler[C]
}

func (d commandTracingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)

	attrs := attribute.String("body", fmt.Sprintf("%#v", cmd))
	ctx, span := otel.Tracer("cmd").Start(ctx, handlerType, trace.WithAttributes(attrs))
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type queryTracingDecorator[C any, R any] struct {
	base QueryHandler[C, R]
}

func (d queryTracingDecorator[C, R]) Handle(ctx context.Context, query C) (result R, err error) {
	handlerType := generateActionName(query)

	attrs := attribute.String("body", fmt.Sprintf("%#v", query))
	ctx, span := otel.Tracer("query").Start(ctx, handlerType, trace.WithAttributes(attrs))
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()
	return d.base.Handle(ctx, query)
}
