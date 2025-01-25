package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type CoinQuery interface {
	ListCoinsQuery
	GetCoinQuery
}

type GetCoinQuery interface {
	GetCoin(ctx context.Context, coinID domain.CoinID) (*domain.Coin, error)
}

type ListCoinsQuery interface {
	ListCoins(ctx context.Context) ([]*domain.Coin, error)
}
