package local

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"go.uber.org/zap"
)

var _ bus.Bus = (*Bus)(nil)

type Bus struct {
	logger *zap.Logger
}

func NewBus(logger *zap.Logger) *Bus {
	return &Bus{
		logger: logger,
	}
}

// Publish implements bus.Bus.
func (b *Bus) Publish(ctx context.Context, event domain.Event) {
	b.logger.Info("publish event", zap.Any("event", event.Topic()))
}
