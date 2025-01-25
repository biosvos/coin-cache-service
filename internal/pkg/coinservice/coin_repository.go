package coinservice

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type ListCoinsQuery interface {
	ListCoins(ctx context.Context) ([]*domain.Coin, error)
}

type ListTradesQuery interface {
	ListTrades(ctx context.Context, coinID domain.CoinID) (*domain.Trades, error)
}

type CoinService interface {
	ListCoinsQuery
	ListTradesQuery
}
