package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type BannedCoinQuery interface {
	ListBannedCoinsQuery
	GetBannedCoinQuery
}

type ListBannedCoinsQuery interface {
	ListBannedCoins(ctx context.Context) ([]*domain.BannedCoin, error)
}

type GetBannedCoinQuery interface {
	GetBannedCoin(ctx context.Context, coinID domain.CoinID) (*domain.BannedCoin, error)
}
