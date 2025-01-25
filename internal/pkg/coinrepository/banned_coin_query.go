package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type BannedCoinQuery interface {
	ListBannedCoinsQuery
}

type ListBannedCoinsQuery interface {
	ListBannedCoins(ctx context.Context) ([]*domain.BannedCoin, error)
}
