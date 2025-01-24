package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type CoinQuery interface {
	ListCoinsQuery
}

type ListCoinsQuery interface {
	ListCoins(ctx context.Context) ([]*domain.Coin, error)
}
