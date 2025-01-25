package flow

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	setpkg "github.com/biosvos/coin-cache-service/internal/pkg/set"
)

type Repository interface {
	coinrepository.ListCoinsQuery
	coinrepository.ListBannedCoinsQuery
	coinrepository.ListTradesQuery
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListCoins(ctx context.Context) ([]string, error) {
	bannedCoins, err := s.repo.ListBannedCoins(ctx)
	if err != nil {
		return nil, err
	}
	bannedCoinSet := setpkg.NewSet(func(coin *domain.BannedCoin) domain.CoinID {
		return coin.CoinID()
	})
	bannedCoinSet.Add(bannedCoins...)
	coins, err := s.repo.ListCoins(ctx)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, coin := range coins {
		if bannedCoinSet.ContainKey(coin.ID()) {
			continue
		}
		ret = append(ret, string(coin.ID()))
	}
	return ret, nil
}

func (s *Service) ListTrades(ctx context.Context, coinID domain.CoinID) (*domain.Trades, error) {
	trades, err := s.repo.ListTrades(ctx, coinID)
	if err != nil {
		return nil, err
	}
	return trades, nil
}
