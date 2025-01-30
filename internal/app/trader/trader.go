package trader

import (
	"context"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/pkg/tracer"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	coinrepository.ListTradesQuery
}

type Repository interface {
	coinrepository.ListBannedCoinsQuery
	coinrepository.ListCoinsQuery
	coinrepository.SaveTradesCommand
}

type Trader struct {
	bus       bus.Bus
	service   Service
	repo      Repository
	scheduler gocron.Scheduler
	logger    *zap.Logger
	jobMap    map[domain.CoinID]uuid.UUID
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
		jobMap:    make(map[domain.CoinID]uuid.UUID),
	}

	for _, coin := range coins {
		if _, ok := bannedCoinMap[coin.ID()]; ok {
			continue
		}
		job, _ := ret.scheduler.NewJob(
			gocron.DurationJob(interval),
			gocron.NewTask(
				ret.RefreshTrades,
				coin.ID(),
			),
		)
		ret.jobMap[coin.ID()] = job.ID()
	}
	return &ret
}

func (t *Trader) Start(ctx context.Context) {
	t.scheduler.Start()
	t.bus.Subscribe(ctx, domain.CoinCreatedEventTopic, func(_ context.Context, event domain.Event) error {
		coinCreatedEvent := domain.ParseCoinCreatedEvent(event.Payload())
		job, _ := t.scheduler.NewJob(
			gocron.DurationJob(interval),
			gocron.NewTask(
				t.RefreshTrades,
				coinCreatedEvent.CoinID,
			),
		)
		t.jobMap[coinCreatedEvent.CoinID] = job.ID()
		err := job.RunNow()
		if err != nil {
			t.logger.Error("failed to run job", zap.Error(err))
		}
		return nil
	})
	t.bus.Subscribe(ctx, domain.CoinDeletedEventTopic, func(_ context.Context, event domain.Event) error {
		coinDeletedEvent := domain.ParseCoinDeletedEvent(event.Payload())
		err := t.scheduler.RemoveJob(t.jobMap[coinDeletedEvent.CoinID])
		if err != nil {
			t.logger.Error("failed to remove job", zap.Error(err))
		}
		delete(t.jobMap, coinDeletedEvent.CoinID)
		return nil
	})
	t.bus.Subscribe(ctx, domain.BannedCoinCreatedEventTopic, func(_ context.Context, event domain.Event) error {
		bannedCoinCreatedEvent := domain.ParseBannedCoinCreatedEvent(event.Payload())
		job, _ := t.scheduler.NewJob(
			gocron.DurationJob(interval),
			gocron.NewTask(
				t.RefreshTrades,
				bannedCoinCreatedEvent.CoinID,
			),
		)
		t.jobMap[bannedCoinCreatedEvent.CoinID] = job.ID()
		err := job.RunNow()
		if err != nil {
			t.logger.Error("failed to run job", zap.Error(err))
		}
		return nil
	})
	t.bus.Subscribe(ctx, domain.BannedCoinDeletedEventTopic, func(_ context.Context, event domain.Event) error {
		bannedCoinDeletedEvent := domain.ParseBannedCoinDeletedEvent(event.Payload())
		err := t.scheduler.RemoveJob(t.jobMap[bannedCoinDeletedEvent.CoinID])
		if err != nil {
			t.logger.Error("failed to remove job", zap.Error(err))
		}
		delete(t.jobMap, bannedCoinDeletedEvent.CoinID)
		return nil
	})
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
