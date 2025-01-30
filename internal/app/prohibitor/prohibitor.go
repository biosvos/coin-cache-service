package prohibitor

import (
	"context"
	"strconv"
	"strings"
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

	coinrepository.ListTradesQuery

	coinrepository.CreateBannedCoinCommand
	coinrepository.ListBannedCoinsQuery
	coinrepository.GetBannedCoinQuery
	coinrepository.DeleteBannedCoinCommand
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
	err := p.checkAndAllowCoins(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	err = p.checkAndProhibitCoins(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	p.bus.Subscribe(ctx, domain.CoinCreatedEventTopic, p.handleCoinCreated)
	p.bus.Subscribe(ctx, domain.CoinUpdatedEventTopic, p.handleCoinUpdated)
	return nil
}

func (p *Prohibitor) checkAndAllowCoins(ctx context.Context) error {
	bannedCoins, err := p.repo.ListBannedCoins(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, bannedCoin := range bannedCoins {
		err := p.checkAndAllowCoin(ctx, bannedCoin.CoinID())
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (p *Prohibitor) checkAndAllowCoin(ctx context.Context, coinID domain.CoinID) error {
	now := time.Now()
	bannedCoin, err := p.repo.GetBannedCoin(ctx, coinID)
	if err != nil {
		return errors.WithStack(err)
	}
	if !bannedCoin.IsBanOver(now) {
		return nil
	}
	err = p.repo.DeleteBannedCoin(ctx, bannedCoin)
	if err != nil {
		return errors.WithStack(err)
	}
	p.logger.Info("allowed coin", zap.String("coin_id", string(bannedCoin.CoinID())))
	p.bus.Publish(ctx, domain.NewBannedCoinDeletedEvent(bannedCoin.CoinID()))
	return nil
}

func (p *Prohibitor) checkAndProhibitCoins(ctx context.Context) error {
	coins, err := p.repo.ListCoins(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, coin := range coins {
		err := p.CheckAndProhibitCoinByStatus(ctx, coin.ID())
		if err != nil {
			return errors.WithStack(err)
		}
		err = p.CheckAndProhibitCoinByTrades(ctx, coin.ID())
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (p *Prohibitor) handleCoinCreated(ctx context.Context, event domain.Event) error {
	coinCreatedEvent := domain.ParseCoinCreatedEvent(event.Payload())
	return p.CheckAndProhibitCoinByStatus(ctx, coinCreatedEvent.CoinID)
}

func (p *Prohibitor) handleCoinUpdated(ctx context.Context, event domain.Event) error {
	coinUpdatedEvent := domain.ParseCoinUpdatedEvent(event.Payload())
	return p.CheckAndProhibitCoinByStatus(ctx, coinUpdatedEvent.CoinID)
}

const day = 24 * time.Hour

func (p *Prohibitor) CheckAndProhibitCoinByStatus(ctx context.Context, coinID domain.CoinID) error {
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

	bannedCoin := domain.NewBannedCoin(coinID, time.Now(), day)

	_, err = p.repo.CreateBannedCoin(ctx, bannedCoin)
	if err != nil {
		return errors.WithStack(err)
	}
	p.bus.Publish(ctx, domain.NewBannedCoinCreatedEvent(coinID))
	p.logger.Info("prohibited coin", zap.String("coin_id", string(coinID)))
	return nil
}

func (p *Prohibitor) CheckAndProhibitCoinByTrades(ctx context.Context, coinID domain.CoinID) error {
	alreadyBanned, err := p.isAlreadyBanned(ctx, coinID)
	if err != nil {
		return errors.WithStack(err)
	}
	if alreadyBanned {
		return nil
	}

	trades, err := p.repo.ListTrades(ctx, coinID)
	if err != nil {
		return errors.WithStack(err)
	}

	var banDuration time.Duration
	if !trades.IsEnoughTrade() { // 20개 이하면, 20개 이상이 될때까지 금지한다.
		banDuration = max(banDuration, day)
	}

	lastPrice := trades.LastPrice()
	sp := strings.Split(string(lastPrice), ".")
	price, err := strconv.ParseUint(sp[0], 10, 64)
	if err != nil {
		return errors.WithStack(err)
	}

	const maxPrice = 100_000
	const tenDays = 10 * day
	if maxPrice < price {
		banDuration = max(banDuration, tenDays)
	}

	const minPrice = 100
	if price < minPrice {
		banDuration = max(banDuration, tenDays)
	}

	bannedCoin := domain.NewBannedCoin(coinID, time.Now(), banDuration)

	if banDuration == 0 {
		return nil
	}
	_, err = p.repo.CreateBannedCoin(ctx, bannedCoin)
	if err != nil {
		return errors.WithStack(err)
	}
	p.logger.Info("prohibited coin", zap.String("coin_id", string(coinID)))
	p.bus.Publish(ctx, domain.NewBannedCoinCreatedEvent(coinID))
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
