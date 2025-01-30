//go:build integration

package telemetry_test

import (
	"context"
	"testing"

	"github.com/biosvos/coin-cache-service/pkg/tracer/telemetry"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tracer, err := telemetry.NewTracer(ctx, "127.0.0.1:4318", "test")
	t.Cleanup(func() {
		tracer.Shutdown()
	})

	outer, span := tracer.Start(ctx, "non")
	t.Cleanup(func() {
		span.End()
	})
	span.Bool("outer", true)
	_, inner := tracer.Start(outer, "kow")
	t.Cleanup(func() {
		inner.End()
	})
	inner.Bool("inner", true)

	assert.NoError(t, err)
}
