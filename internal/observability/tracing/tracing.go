// Copyright 2021-2022 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package tracing

import (
	"context"
	"fmt"
	"net/http"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace/propagation"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/cerbos/cerbos/internal/config"
)

var conf Conf

func Init(ctx context.Context) error {
	if err := config.GetSection(&conf); err != nil {
		return fmt.Errorf("failed to load tracing config: %w", err)
	}

	switch conf.Exporter {
	case jaegerExporter:
		return configureJaeger(ctx)
	case "":
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return nil
	default:
		return fmt.Errorf("unknown exporter %q", conf.Exporter)
	}
}

func HTTPHandler(handler http.Handler) http.Handler {
	var prop propagation.HTTPFormat
	if conf.PropagationFormat == propagationW3CTraceContext {
		prop = &tracecontext.HTTPFormat{}
	}

	return &ochttp.Handler{
		Handler:     handler,
		Propagation: prop,
	}
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("cerbos.dev/cerbos").Start(ctx, name)
}

func MarkFailed(span trace.Span, code int, err error) {
	if err != nil {
		span.RecordError(err)
	}

	c, desc := semconv.SpanStatusFromHTTPStatusCode(code)
	span.SetStatus(c, desc)
}

type otelErrHandler func(err error)

func (o otelErrHandler) Handle(err error) {
	o(err)
}
