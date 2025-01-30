package telemetry

import (
	"github.com/biosvos/coin-cache-service/pkg/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ tracer.Span = (*Span)(nil)

type Span struct {
	span trace.Span
}

func newSpan(span trace.Span) *Span {
	return &Span{
		span: span,
	}
}

// AddAttribute implements tracer.Span.
func (s *Span) String(key string, value string) {
	s.span.SetAttributes(attribute.String(key, value))
}

func (s *Span) Int64(key string, value int64) {
	s.span.SetAttributes(attribute.Int64(key, value))
}

func (s *Span) Bool(key string, value bool) {
	s.span.SetAttributes(attribute.Bool(key, value))
}

func (s *Span) Float64(key string, value float64) {
	s.span.SetAttributes(attribute.Float64(key, value))
}

func (s *Span) Error(err error) {
	s.span.RecordError(err)
}

func (s *Span) End() {
	s.span.End()
}
