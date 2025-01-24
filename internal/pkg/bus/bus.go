package bus

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type Bus interface {
	Publish(ctx context.Context, event domain.Event)
}
