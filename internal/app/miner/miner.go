package miner

import (
	"context"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	setpkg "github.com/biosvos/coin-cache-service/internal/pkg/set"
	"github.com/go-co-op/gocron/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Repository interface {
	coinrepository.ListCoinsQuery
	coinrepository.ListBannedCoinsQuery
	coinrepository.CreateCoinCommand
	coinrepository.UpdateCoinCommand
	coinrepository.DeleteCoinCommand
}

type Service interface {
	coinrepository.ListCoinsQuery
}

// Miner coin 정보를 최신화한다.
type Miner struct {
	service    Service
	repository Repository
	bus        bus.Bus
	logger     *zap.Logger

	scheduler gocron.Scheduler
}

func NewMiner(
	logger *zap.Logger,
	service Service,
	repository Repository,
	bus bus.Bus,
) *Miner {
	scheduler, _ := gocron.NewScheduler() // option이 없으면 error도 발생하지 않는다.
	ret := Miner{
		logger:     logger,
		service:    service,
		repository: repository,
		bus:        bus,
		scheduler:  scheduler,
	}
	_, _ = ret.scheduler.NewJob(
		gocron.DurationJob(
			mineInterval,
		),
		gocron.NewTask(
			func() {
				ret.logger.Info("run task")
				defer ret.logger.Info("task done")

				ctx := context.Background()
				err := ret.Mine(ctx)
				if err != nil {
					ret.logger.Error("failed to mine", zap.Error(err))
				}
			},
		),
	)
	return &ret
}

const mineInterval = 10 * time.Minute

func (m *Miner) Start() error {
	m.scheduler.Start()
	for _, job := range m.scheduler.Jobs() {
		err := job.RunNow()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (m *Miner) Stop() {
	err := m.scheduler.Shutdown()
	if err != nil {
		m.logger.Error("failed to shutdown scheduler", zap.Error(err))
	}
}

func (m *Miner) Mine(ctx context.Context) error { //nolint:cyclop  //FIXME 나중에 nolint 제거
	m.logger.Info("start mine")
	defer m.logger.Info("mine done")

	bannedCoinSet, err := m.listBannedCoinSet(ctx)
	if err != nil {
		return err
	}

	repositoryCoins, err := m.listRepositoryCoinsWithoutBanned(ctx, bannedCoinSet)
	if err != nil {
		return err
	}
	now := time.Now()
	if !m.needRefresh(now, repositoryCoins) {
		return nil
	}

	serviceCoins, err := m.listServiceCoinsWithoutBanned(ctx, bannedCoinSet)
	if err != nil {
		return err
	}

	repositoryCoinSet := setpkg.NewSet(func(coin *domain.Coin) domain.CoinID {
		return coin.ID()
	})
	repositoryCoinSet.Add(repositoryCoins...)
	serviceCoinSet := setpkg.NewSet(func(coin *domain.Coin) domain.CoinID {
		return coin.ID()
	})
	serviceCoinSet.Add(serviceCoins...)

	onlyServiceCoinSet := serviceCoinSet.Difference(repositoryCoinSet)
	onlyRepositoryCoinSet := repositoryCoinSet.Difference(serviceCoinSet)
	bothCoinSet := serviceCoinSet.Intersection(repositoryCoinSet)

	for _, coin := range onlyServiceCoinSet.Values() {
		err := m.createCoin(ctx, now, coin)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	for _, coin := range onlyRepositoryCoinSet.Values() {
		err := m.deleteCoin(ctx, now, coin)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	for _, coin := range bothCoinSet.Values() {
		if !coin.IsOld(now) {
			continue
		}
		coin = coin.SetModifiedAt(now)
		err := m.updateCoin(ctx, now, coin)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (m *Miner) needRefresh(now time.Time, repositoryCoins []*domain.Coin) bool {
	if len(repositoryCoins) == 0 {
		return true // 코인이 없다면 최신화가 필요하다.
	}
	for _, coin := range repositoryCoins {
		if coin.IsOld(now) {
			return true
		}
	}
	return false
}

func (m *Miner) listBannedCoinSet(ctx context.Context) (*setpkg.Set[domain.CoinID, *domain.BannedCoin], error) {
	bannedCoins, err := m.repository.ListBannedCoins(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	bannedCoinSet := setpkg.NewSet(
		func(bannedCoin *domain.BannedCoin) domain.CoinID {
			return bannedCoin.CoinID()
		},
	)
	bannedCoinSet.Add(bannedCoins...)
	return bannedCoinSet, nil
}

func (m *Miner) listServiceCoinsWithoutBanned(
	ctx context.Context,
	bannedCoinSet *setpkg.Set[domain.CoinID, *domain.BannedCoin],
) ([]*domain.Coin, error) {
	serviceCoins, err := m.service.ListCoins(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var filteredCoins []*domain.Coin
	for _, coin := range serviceCoins {
		if bannedCoinSet.ContainKey(coin.ID()) {
			continue
		}
		filteredCoins = append(filteredCoins, coin)
	}
	return filteredCoins, nil
}

func (m *Miner) listRepositoryCoinsWithoutBanned(
	ctx context.Context,
	bannedCoinSet *setpkg.Set[domain.CoinID, *domain.BannedCoin],
) ([]*domain.Coin, error) {
	repositoryCoins, err := m.repository.ListCoins(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var filteredCoins []*domain.Coin
	for _, coin := range repositoryCoins {
		if bannedCoinSet.ContainKey(coin.ID()) {
			continue
		}
		filteredCoins = append(filteredCoins, coin)
	}
	return filteredCoins, nil
}

func (m *Miner) createCoin(ctx context.Context, now time.Time, coin *domain.Coin) error {
	_, err := m.repository.CreateCoin(ctx, coin)
	if err != nil {
		return errors.WithStack(err)
	}
	m.logger.Info("create coin", zap.String("coin_id", string(coin.ID())))
	m.bus.Publish(ctx, domain.NewCoinCreatedEvent(now, coin.ID()))
	return nil
}

func (m *Miner) deleteCoin(ctx context.Context, now time.Time, coin *domain.Coin) error {
	err := m.repository.DeleteCoin(ctx, coin)
	if err != nil {
		return errors.WithStack(err)
	}
	m.logger.Info("delete coin", zap.String("coin_id", string(coin.ID())))
	m.bus.Publish(ctx, domain.NewCoinDeletedEvent(now, coin.ID()))
	return nil
}

func (m *Miner) updateCoin(ctx context.Context, now time.Time, coin *domain.Coin) error {
	_, err := m.repository.UpdateCoin(ctx, coin)
	if err != nil {
		return errors.WithStack(err)
	}
	m.logger.Info("update coin", zap.String("coin_id", string(coin.ID())))
	m.bus.Publish(ctx, domain.NewCoinUpdatedEvent(now, coin.ID()))
	return nil
}
