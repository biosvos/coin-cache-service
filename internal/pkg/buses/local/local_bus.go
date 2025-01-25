package local

import (
	"context"
	"errors"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"go.uber.org/zap"
)

var _ bus.Bus = (*Bus)(nil)

type EventHandler func(ctx context.Context, event domain.Event) error

type Bus struct {
	logger   *zap.Logger
	handlers map[string][]EventHandler
}

func NewBus(logger *zap.Logger) *Bus {
	return &Bus{
		logger:   logger,
		handlers: make(map[string][]EventHandler),
	}
}

// Publish implements bus.Bus.
func (b *Bus) Publish(ctx context.Context, event domain.Event) {
	b.logger.Info("publish event", zap.Any("event", event.Topic()))
	for _, handler := range b.handlers[event.Topic()] {
		b.handle(ctx, handler, event)
	}
}

func (b *Bus) handle(ctx context.Context, handler EventHandler, event domain.Event) {
	for {
		err := handler(ctx, event)
		if err == nil {
			return
		}
		var retryAfterError *bus.RetryAfterError
		if errors.As(err, &retryAfterError) {
			b.logger.Info("retry after", zap.Duration("duration", retryAfterError.Duration()))
			time.Sleep(retryAfterError.Duration())
			continue
		}
		b.logger.Error("error handling event", zap.Error(err))
		return
	}
}

func (b *Bus) Subscribe(_ context.Context, topic string, handler func(ctx context.Context, event domain.Event) error) {
	b.handlers[topic] = append(b.handlers[topic], handler)
}
