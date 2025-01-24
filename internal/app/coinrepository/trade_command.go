package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type TradeCommand interface {
	CreateTradeCommand
	UpdateTradeCommand
	DeleteTradeCommand
}

type CreateTradeCommand interface {
	CreateTrade(ctx context.Context, trade *domain.Trade) (*domain.Trade, error)
}

type UpdateTradeCommand interface {
	UpdateTrade(ctx context.Context, trade *domain.Trade) (*domain.Trade, error)
}

type DeleteTradeCommand interface {
	DeleteTrade(ctx context.Context, id domain.CoinID) error
}
