package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type TradeCommand interface {
	SaveTradesCommand
	DeleteTradesCommand
}

type SaveTradesCommand interface {
	SaveTrades(ctx context.Context, id domain.CoinID, trades []*domain.Trade) error
}

type DeleteTradesCommand interface {
	DeleteTrades(ctx context.Context, id domain.CoinID) error
}
