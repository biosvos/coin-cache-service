package prohibitor

import (
	"context"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Repository interface {
	coinrepository.ListCoinsQuery
	coinrepository.GetCoinQuery

	coinrepository.CreateBannedCoinCommand
	coinrepository.GetBannedCoinQuery
}

type Prohibitor struct {
	logger *zap.Logger
	bus    bus.Bus
	repo   Repository
}

func NewProhibitor(logger *zap.Logger, bus bus.Bus, repo Repository) *Prohibitor {
	return &Prohibitor{logger: logger, bus: bus, repo: repo}
}

func (p *Prohibitor) Start(ctx context.Context) error {
	coins, err := p.repo.ListCoins(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, coin := range coins {
		err := p.ProhibitCoin(ctx, coin.ID())
		if err != nil {
			return errors.WithStack(err)
		}
	}
	p.bus.Subscribe(ctx, domain.CoinCreatedEventTopic, p.handleCoinCreated)
	p.bus.Subscribe(ctx, domain.CoinUpdatedEventTopic, p.handleCoinUpdated)
	return nil
}

func (p *Prohibitor) handleCoinCreated(ctx context.Context, event domain.Event) error {
	coinCreatedEvent := domain.ParseCoinCreatedEvent(event.Payload())
	return p.ProhibitCoin(ctx, coinCreatedEvent.CoinID)
}

func (p *Prohibitor) handleCoinUpdated(ctx context.Context, event domain.Event) error {
	coinUpdatedEvent := domain.ParseCoinUpdatedEvent(event.Payload())
	return p.ProhibitCoin(ctx, coinUpdatedEvent.CoinID)
}

func (p *Prohibitor) ProhibitCoin(ctx context.Context, coinID domain.CoinID) error {
	alreadyBanned, err := p.isAlreadyBanned(ctx, coinID)
	if err != nil {
		return errors.WithStack(err)
	}
	if alreadyBanned {
		return nil
	}

	coin, err := p.repo.GetCoin(ctx, coinID)
	if err != nil {
		return errors.WithStack(err)
	}
	if !coin.IsDanger() {
		return nil
	}

	const day = 24 * time.Hour
	bannedCoin := domain.NewBannedCoin(coinID, time.Now(), day)

	_, err = p.repo.CreateBannedCoin(ctx, bannedCoin)
	if err != nil {
		return errors.WithStack(err)
	}
	p.logger.Info("prohibited coin", zap.String("coin_id", string(coinID)))
	return nil
}

func (p *Prohibitor) isAlreadyBanned(ctx context.Context, coinID domain.CoinID) (bool, error) {
	_, err := p.repo.GetBannedCoin(ctx, coinID)
	if err != nil {
		if errors.Is(err, coinrepository.ErrBannedCoinNotFound) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}
