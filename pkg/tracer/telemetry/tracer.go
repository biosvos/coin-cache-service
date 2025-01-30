package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/biosvos/coin-cache-service/pkg/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	tracecontext "go.opentelemetry.io/otel/trace"
)

var _ tracer.Tracer = (*Tracer)(nil)

type Tracer struct {
	provider *trace.TracerProvider
}

func NewTracer(ctx context.Context, serviceName string, opts ...Option) (*Tracer, error) {
	options := NewOptions()
	for _, opt := range opts {
		opt(options)
	}
	exporter, err := otlptrace.New(
		ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(options.URL),
			otlptracehttp.WithHeaders(map[string]string{
				"Content-Type": "application/json",
			}),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter: %w", err)
	}
	stdExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating stdout exporter: %w", err)
	}

	tracerprovider := trace.NewTracerProvider(
		trace.WithBatcher(
			exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithBatcher(
			stdExporter,
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			),
		),
	)
	otel.SetTracerProvider(tracerprovider)
	return &Tracer{
		provider: tracerprovider,
	}, nil
}

func (t *Tracer) Start(ctx context.Context, name string) (context.Context, tracer.Span) {
	if !tracecontext.SpanContextFromContext(ctx).IsValid() {
		tracer := t.provider.Tracer(name)
		ctx, span := tracer.Start(ctx, name) //nolint:spancheck
		return ctx, newSpan(span)            //nolint:spancheck
	}
	tracer := t.provider.Tracer(name)
	ctx, span := tracer.Start(ctx, name) //nolint:spancheck
	return ctx, newSpan(span)            //nolint:spancheck
}

// Shutdown implements tracer.Tracer.
func (t *Tracer) Shutdown() {
	err := t.provider.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}
