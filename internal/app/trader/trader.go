package trader

import (
	"context"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinservice"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Service interface {
	coinservice.ListTradesQuery
}

type Repository interface {
	coinrepository.ListBannedCoinsQuery
	coinrepository.SaveTradesCommand
}

type Trader struct {
	bus     bus.Bus
	service Service
	repo    Repository
	logger  *zap.Logger
}

func NewTrader(logger *zap.Logger, bus bus.Bus, service Service, repo Repository) *Trader {
	return &Trader{logger: logger, bus: bus, service: service, repo: repo}
}

func (t *Trader) Start(ctx context.Context) {
	t.bus.Subscribe(ctx, domain.CoinCreatedEventTopic, func(ctx context.Context, event domain.Event) error {
		coinCreatedEvent := domain.ParseCoinCreatedEvent(event.Payload())
		return t.RefreshTrades(ctx, coinCreatedEvent.CoinID)
	})
}

func (t *Trader) RefreshTrades(ctx context.Context, coinID domain.CoinID) error {
	t.logger.Info("refreshing trades", zap.String("coin_id", string(coinID)))
	trades, err := t.service.ListTrades(ctx, coinID)
	if err != nil {
		return errors.WithStack(err)
	}
	err = t.repo.SaveTrades(ctx, trades)
	if err != nil {
		return errors.WithStack(err)
	}
	t.logger.Info("refreshed trades", zap.String("coin_id", string(coinID)))
	return nil
}
