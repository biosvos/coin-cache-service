package prohibitor

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/go-co-op/gocron/v2"
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
	logger    *zap.Logger
	bus       bus.Bus
	repo      Repository
	scheduler gocron.Scheduler
}

const day = 24 * time.Hour

func NewProhibitor(logger *zap.Logger, bus bus.Bus, repo Repository) *Prohibitor {
	scheduler, _ := gocron.NewScheduler()
	return &Prohibitor{logger: logger, bus: bus, repo: repo, scheduler: scheduler}
}

func (p *Prohibitor) Start(ctx context.Context) error {
	p.scheduler.Start()
	p.bus.Subscribe(ctx, domain.CoinCreatedEventTopic, p.handleCoinCreated)
	p.bus.Subscribe(ctx, domain.CoinUpdatedEventTopic, p.handleCoinUpdated)
	p.bus.Subscribe(ctx, domain.CoinDeletedEventTopic, p.handleCoinDeleted)
	p.bus.Subscribe(ctx, domain.TradesUpdatedEventTopic, p.handleTradesUpdated)
	p.bus.Subscribe(ctx, domain.TradesDeletedEventTopic, p.handleTradesDeleted)
	return nil
}

func (p *Prohibitor) Stop() {
	err := p.scheduler.Shutdown()
	if err != nil {
		p.logger.Error("failed to shutdown scheduler", zap.Error(err))
	}
}

func (p *Prohibitor) addExpireBannedCoinJob(ctx context.Context, bannedCoin *domain.BannedCoin) {
	_, err := p.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(
				bannedCoin.ExpiredAt(),
			),
		),
		gocron.NewTask(
			func(coinID domain.CoinID) {
				bannedCoin, err := p.repo.GetBannedCoin(ctx, coinID)
				if err != nil {
					p.logger.Error("failed to get banned coin", zap.Error(err))
					return
				}
				err = p.deleteBannedCoin(ctx, bannedCoin)
				if err != nil {
					p.logger.Error("failed to delete banned coin", zap.Error(err))
				}
			},
			bannedCoin.CoinID(),
		),
	)
	if err != nil {
		p.logger.Error("failed to add expire banned coin job", zap.Error(err))
	}
}

func (p *Prohibitor) handleCoinCreated(ctx context.Context, event domain.Event) error {
	coinCreatedEvent := domain.ParseCoinCreatedEvent(event.Payload())
	return p.prohibitByStatus(ctx, coinCreatedEvent.CoinID)
}

func (p *Prohibitor) handleCoinUpdated(ctx context.Context, event domain.Event) error {
	coinUpdatedEvent := domain.ParseCoinUpdatedEvent(event.Payload())
	return p.prohibitByStatus(ctx, coinUpdatedEvent.CoinID)
}

func (p *Prohibitor) handleCoinDeleted(ctx context.Context, event domain.Event) error {
	coinDeletedEvent := domain.ParseCoinDeletedEvent(event.Payload())
	return p.allowCoin(ctx, coinDeletedEvent.CoinID)
}

func (p *Prohibitor) handleTradesUpdated(ctx context.Context, event domain.Event) error {
	tradesUpdatedEvent := domain.ParseTradesUpdatedEvent(event.Payload())
	return p.prohibitByTrades(ctx, tradesUpdatedEvent.CoinID)
}

func (p *Prohibitor) handleTradesDeleted(ctx context.Context, event domain.Event) error {
	tradesDeletedEvent := domain.ParseTradesDeletedEvent(event.Payload())
	return p.allowCoin(ctx, tradesDeletedEvent.CoinID)
}

func (p *Prohibitor) prohibitByStatus(ctx context.Context, coinID domain.CoinID) error {
	_, err := p.repo.GetBannedCoin(ctx, coinID)
	if err == nil { // already banned
		return nil
	}
	if !errors.Is(err, coinrepository.ErrBannedCoinNotFound) {
		return errors.WithStack(err)
	}
	coin, err := p.repo.GetCoin(ctx, coinID)
	if err != nil {
		return errors.WithStack(err)
	}
	if !coin.IsDanger() {
		return nil
	}
	err = p.createBannedCoin(ctx, coinID, day)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *Prohibitor) prohibitByTrades(ctx context.Context, coinID domain.CoinID) error {
	_, err := p.repo.GetBannedCoin(ctx, coinID)
	if err == nil { // already banned
		return nil
	}
	if !errors.Is(err, coinrepository.ErrBannedCoinNotFound) {
		return errors.WithStack(err)
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

	if banDuration == 0 {
		return nil
	}
	err = p.createBannedCoin(ctx, coinID, banDuration)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *Prohibitor) allowCoin(ctx context.Context, coinID domain.CoinID) error {
	bannedCoin, err := p.repo.GetBannedCoin(ctx, coinID)
	if err != nil {
		if errors.Is(err, coinrepository.ErrBannedCoinNotFound) {
			// already allowed
			return nil
		}
		return errors.WithStack(err)
	}
	if !bannedCoin.IsBanOver(time.Now()) {
		// still banned
		return nil
	}
	return p.deleteBannedCoin(ctx, bannedCoin)
}

func (p *Prohibitor) createBannedCoin(ctx context.Context, coinID domain.CoinID, period time.Duration) error {
	bannedCoin := domain.NewBannedCoin(coinID, time.Now(), period)
	_, err := p.repo.CreateBannedCoin(ctx, bannedCoin)
	if err != nil {
		return errors.WithStack(err)
	}
	p.logger.Info("prohibited coin", zap.String("coin_id", string(coinID)))
	p.bus.Publish(ctx, domain.NewBannedCoinCreatedEvent(coinID))
	p.addExpireBannedCoinJob(ctx, bannedCoin)
	return nil
}

func (p *Prohibitor) deleteBannedCoin(ctx context.Context, bannedCoin *domain.BannedCoin) error {
	err := p.repo.DeleteBannedCoin(ctx, bannedCoin)
	if err != nil {
		return errors.WithStack(err)
	}
	p.logger.Info("allowed coin", zap.String("coin_id", string(bannedCoin.CoinID())))
	p.bus.Publish(ctx, domain.NewBannedCoinDeletedEvent(bannedCoin.CoinID()))
	return nil
}
