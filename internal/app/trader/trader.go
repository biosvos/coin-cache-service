package trader

import (
	"context"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/pkg/tracer"
	"github.com/go-co-op/gocron/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Service interface {
	coinrepository.ListTradesQuery
}

type Repository interface {
	coinrepository.ListBannedCoinsQuery
	coinrepository.ListCoinsQuery
	coinrepository.SaveTradesCommand
	coinrepository.DeleteTradesCommand
	coinrepository.GetCoinQuery
}

type Trader struct {
	bus       bus.Bus
	service   Service
	repo      Repository
	scheduler gocron.Scheduler
	logger    *zap.Logger
	tracer    tracer.Tracer
}

const interval = time.Minute * 10

func NewTrader(tracer tracer.Tracer, logger *zap.Logger, bus bus.Bus, service Service, repo Repository) *Trader {
	scheduler, _ := gocron.NewScheduler() // option이 없으면 error도 발생하지 않는다.
	ctx := context.Background()
	coins, err := repo.ListCoins(ctx)
	if err != nil {
		panic(err)
	}
	bannedCoins, err := repo.ListBannedCoins(ctx)
	if err != nil {
		panic(err)
	}
	bannedCoinMap := make(map[domain.CoinID]struct{})
	for _, coin := range bannedCoins {
		bannedCoinMap[coin.CoinID()] = struct{}{}
	}
	ret := Trader{
		tracer:    tracer,
		logger:    logger,
		bus:       bus,
		service:   service,
		repo:      repo,
		scheduler: scheduler,
	}

	for _, coin := range coins {
		if _, ok := bannedCoinMap[coin.ID()]; ok {
			continue
		}
		_ = ret.addRefreshTradesJob(coin.ID())
	}
	return &ret
}

func (t *Trader) Start(ctx context.Context) {
	t.scheduler.Start()
	t.bus.Subscribe(ctx, domain.CoinCreatedEventTopic, t.handleCoinCreatedEvent)
	t.bus.Subscribe(ctx, domain.CoinDeletedEventTopic, t.handleCoinDeletedEvent)
	t.bus.Subscribe(ctx, domain.BannedCoinCreatedEventTopic, t.handleBannedCoinCreatedEvent)
	t.bus.Subscribe(ctx, domain.BannedCoinDeletedEventTopic, t.handleBannedCoinDeletedEvent)
	for _, job := range t.scheduler.Jobs() {
		err := job.RunNow()
		if err != nil {
			t.logger.Error("failed to run job", zap.Error(err))
		}
	}
}

func (t *Trader) Stop() {
	err := t.scheduler.Shutdown()
	if err != nil {
		t.logger.Error("failed to shutdown scheduler", zap.Error(err))
	}
}

func (t *Trader) addRefreshTradesJob(coinID domain.CoinID) gocron.Job {
	job, _ := t.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(
			t.RefreshTrades,
			coinID,
		),
		gocron.WithTags(string(coinID)),
	)
	return job
}

func (t *Trader) removeRefreshTradesJob(coinID domain.CoinID) {
	t.scheduler.RemoveByTags(string(coinID))
}

func (t *Trader) handleCoinCreatedEvent(ctx context.Context, event domain.Event) error {
	_, span := t.tracer.Start(ctx, "trader.handleCoinCreatedEvent")
	defer span.End()

	coinCreatedEvent := domain.ParseCoinCreatedEvent(event.Payload())
	span.String("coin_id", string(coinCreatedEvent.CoinID))
	job := t.addRefreshTradesJob(coinCreatedEvent.CoinID)
	err := job.RunNow()
	if err != nil {
		span.Error(err)
	}
	return nil
}

func (t *Trader) handleCoinDeletedEvent(ctx context.Context, event domain.Event) error {
	_, span := t.tracer.Start(ctx, "trader.handleCoinDeletedEvent")
	defer span.End()

	coinDeletedEvent := domain.ParseCoinDeletedEvent(event.Payload())
	span.String("coin_id", string(coinDeletedEvent.CoinID))
	t.removeRefreshTradesJob(coinDeletedEvent.CoinID)
	return nil
}

func (t *Trader) handleBannedCoinCreatedEvent(ctx context.Context, event domain.Event) error {
	_, span := t.tracer.Start(ctx, "trader.handleBannedCoinCreatedEvent")
	defer span.End()

	bannedCoinCreatedEvent := domain.ParseBannedCoinCreatedEvent(event.Payload())
	span.String("coin_id", string(bannedCoinCreatedEvent.CoinID))

	t.removeRefreshTradesJob(bannedCoinCreatedEvent.CoinID)
	return nil
}

func (t *Trader) handleBannedCoinDeletedEvent(ctx context.Context, event domain.Event) error {
	_, span := t.tracer.Start(ctx, "trader.handleBannedCoinDeletedEvent")
	defer span.End()

	bannedCoinDeletedEvent := domain.ParseBannedCoinDeletedEvent(event.Payload())
	span.String("coin_id", string(bannedCoinDeletedEvent.CoinID))

	coin, err := t.repo.GetCoin(ctx, bannedCoinDeletedEvent.CoinID)
	switch {
	case errors.Is(err, coinrepository.ErrCoinNotFound):
		return nil

	case err == nil:
		job := t.addRefreshTradesJob(coin.ID())
		err = job.RunNow()
		if err != nil {
			span.Error(err)
		}
		return nil

	default:
		span.Error(err)
		return errors.WithStack(err)
	}
}

func (t *Trader) RefreshTrades(coinID domain.CoinID) {
	ctx := context.Background()
	ctx, span := t.tracer.Start(ctx, "trader.RefreshTrades")
	defer span.End()
	span.String("coin_id", string(coinID))

	trades, err := t.service.ListTrades(ctx, coinID)
	if err != nil {
		span.Error(err)
		return
	}
	err = t.repo.SaveTrades(ctx, trades)
	if err != nil {
		span.Error(err)
		return
	}
}

func (t *Trader) DeleteTrades(coinID domain.CoinID) {
	ctx := context.Background()
	ctx, span := t.tracer.Start(ctx, "trader.DeleteTrades")
	defer span.End()
	span.String("coin_id", string(coinID))

	err := t.repo.DeleteTrades(ctx, coinID)
	if err != nil {
		span.Error(err)
	}
}
