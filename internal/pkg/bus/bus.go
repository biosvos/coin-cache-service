package bus

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type Bus interface {
	Publish(ctx context.Context, event domain.Event)
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, event domain.Event) error)
}
