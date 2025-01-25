package coinrepository

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type BannedCoinCommand interface {
	CreateBannedCoinCommand
	UpdateBannedCoinCommand
	DeleteBannedCoinCommand
}

type CreateBannedCoinCommand interface {
	CreateBannedCoin(ctx context.Context, bannedCoin *domain.BannedCoin) (*domain.BannedCoin, error)
}

type UpdateBannedCoinCommand interface {
	UpdateBannedCoin(ctx context.Context, bannedCoin *domain.BannedCoin) (*domain.BannedCoin, error)
}

type DeleteBannedCoinCommand interface {
	DeleteBannedCoin(ctx context.Context, bannedCoin *domain.BannedCoin) error
}
