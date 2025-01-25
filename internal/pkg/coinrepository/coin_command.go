package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type CoinCommand interface {
	CreateCoinCommand
	UpdateCoinCommand
	DeleteCoinCommand
}

type CreateCoinCommand interface {
	CreateCoin(ctx context.Context, coin *domain.Coin) (*domain.Coin, error)
}

type UpdateCoinCommand interface {
	UpdateCoin(ctx context.Context, coin *domain.Coin) (*domain.Coin, error)
}

type DeleteCoinCommand interface {
	DeleteCoin(ctx context.Context, coin *domain.Coin) error
}
