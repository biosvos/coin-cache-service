package tracer

import "context"

type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
	Shutdown()
}

type Span interface {
	End()

	String(key string, value string)
	Int64(key string, value int64)
	Bool(key string, value bool)
	Float64(key string, value float64)
	Error(err error)
}
