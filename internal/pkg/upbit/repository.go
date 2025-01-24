package upbit

import (
	"context"
	"strings"
	"time"

	"github.com/biosvos/coin-cache-service/internal/app/coinservice"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

var _ coinservice.CoinService = (*Service)(nil)

type Service struct {
	upbit *Upbit
}

func NewService() *Service {
	upbit := NewUpbit()
	return &Service{upbit: upbit}
}

func (s *Service) ListCoins(ctx context.Context) ([]*domain.Coin, error) {
	now := time.Now()
	coins, err := s.upbit.ListCoins(ctx)
	if err != nil {
		return nil, err
	}
	var ret []*domain.Coin
	for _, coin := range coins {
		if !strings.HasPrefix(coin.Market, "KRW-") {
			// 한국 화폐만 허용한다.
			continue
		}
		ret = append(ret, coin.ToDomain(now))
	}
	return ret, nil
}
