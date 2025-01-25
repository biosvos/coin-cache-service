package miner

import (
	"context"
	"sync"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/bus"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/coinservice"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	setpkg "github.com/biosvos/coin-cache-service/internal/pkg/set"
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
	coinservice.ListCoinsQuery
}

// Miner coin 정보를 최신화한다.
type Miner struct {
	service    Service
	repository Repository
	bus        bus.Bus
	logger     *zap.Logger

	timer         *time.Timer
	wg            sync.WaitGroup
	isRunningFlag bool
	stopCh        chan struct{}
	mu            sync.Mutex
}

func NewMiner(
	logger *zap.Logger,
	service Service,
	repository Repository,
	bus bus.Bus,
) *Miner {
	return &Miner{
		logger:     logger,
		service:    service,
		repository: repository,
		bus:        bus,

		stopCh: make(chan struct{}),

		wg:            sync.WaitGroup{},
		mu:            sync.Mutex{},
		timer:         nil,
		isRunningFlag: false,
	}
}

const mineInterval = 10 * time.Minute

func (m *Miner) setRunning() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isRunningFlag {
		return errors.New("miner is already running")
	}
	m.isRunningFlag = true
	m.timer = time.NewTimer(mineInterval)
	m.wg.Add(1)
	return nil
}

func (m *Miner) clearRunning() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.wg.Done()
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	m.isRunningFlag = false
}

func (m *Miner) Start(ctx context.Context) error {
	err := m.Mine(ctx)
	if err != nil {
		return err
	}
	err = m.setRunning()
	if err != nil {
		return err
	}
	go func() {
		defer m.clearRunning()
		for {
			select {
			case <-m.timer.C:
				err := m.Mine(ctx)
				if err != nil {
					m.logger.Error("failed to mine", zap.Error(err))
				}
				_ = m.timer.Reset(mineInterval)
			case <-m.stopCh:
				return
			}
		}
	}()
	return nil
}

func (m *Miner) Stop() {
	m.stopCh <- struct{}{}
	m.wg.Wait()
}

func (m *Miner) Mine(ctx context.Context) error {
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
	if m.needRefresh(now, repositoryCoins) {
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
