package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type ListTradesQuery interface {
	ListTrades(ctx context.Context, id domain.CoinID) (*domain.Trades, error)
}
