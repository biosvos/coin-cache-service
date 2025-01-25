package local

import (
	"context"
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
		switch v := err.(type) {
		case *bus.RetryAfterError:
			b.logger.Info("retry after", zap.Duration("duration", v.Duration()))
			time.Sleep(v.Duration())

		default:
			b.logger.Error("error handling event", zap.Error(err))
			return
		}
	}
}

func (b *Bus) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, event domain.Event) error) {
	b.handlers[topic] = append(b.handlers[topic], handler)
}
