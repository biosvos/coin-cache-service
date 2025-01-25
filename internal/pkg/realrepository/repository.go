package realrepository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/internal/pkg/keyvalue"
	"github.com/biosvos/coin-cache-service/internal/pkg/keyvalues/badger"
	"github.com/pkg/errors"
)

var _ coinrepository.CoinRepository = (*Repository)(nil)

type Repository struct {
	kv *badger.Store
}

func NewRepository(path string) *Repository {
	kv, err := badger.NewStore(path)
	if err != nil {
		log.Fatal(err)
	}
	return &Repository{kv: kv}
}

func (r *Repository) Close() {
	r.kv.Close()
}

// CreateCoin implements coinrepository.CoinRepository.
func (r *Repository) CreateCoin(_ context.Context, domainCoin *domain.Coin) (*domain.Coin, error) {
	coin := NewCoin(domainCoin)
	err := r.kv.Create(coin.Key(), coin.Value())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return domainCoin, nil
}

// ListCoins implements coinrepository.CoinRepository.
func (r *Repository) ListCoins(_ context.Context) ([]*domain.Coin, error) {
	var coins []*Coin
	items, err := r.kv.List([]byte(coinPrefix))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, item := range items {
		var coin Coin
		err := json.Unmarshal(item, &coin)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		coins = append(coins, &coin)
	}
	var ret []*domain.Coin
	for _, coin := range coins {
		ret = append(ret, coin.ToDomain())
	}
	return ret, nil
}

// GetCoin implements coinrepository.CoinRepository.
func (r *Repository) GetCoin(_ context.Context, coinID domain.CoinID) (*domain.Coin, error) {
	key := CoinKey(coinID)
	item, err := r.kv.Get(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret Coin
	err = json.Unmarshal(item, &ret)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret.ToDomain(), nil
}

// UpdateCoin implements coinrepository.CoinRepository.
func (r *Repository) UpdateCoin(_ context.Context, domainCoin *domain.Coin) (*domain.Coin, error) {
	coin := NewCoin(domainCoin)
	err := r.kv.Update(coin.Key(), coin.Value())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return domainCoin, nil
}

// DeleteCoin implements coinrepository.CoinRepository.
func (r *Repository) DeleteCoin(_ context.Context, domainCoin *domain.Coin) error {
	coin := NewCoin(domainCoin)
	err := r.kv.Delete(coin.Key())
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SaveTrades implements coinrepository.CoinRepository.
func (r *Repository) SaveTrades(_ context.Context, domainTrades *domain.Trades) error {
	trades := NewTrades(domainTrades.CoinID(), domainTrades.ModifiedAt(), domainTrades.Trades())
	err := r.kv.Update(trades.Key(), trades.Value())
	if err != nil {
		if errors.Is(err, keyvalue.ErrKeyNotFound) {
			err := r.kv.Create(trades.Key(), trades.Value())
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}
		return errors.WithStack(err)
	}
	return nil
}

// ListTrades implements coinrepository.CoinRepository.
func (r *Repository) ListTrades(_ context.Context, id domain.CoinID) (*domain.Trades, error) {
	key := NewTrades(id, time.Time{}, nil)
	item, err := r.kv.Get(key.Key())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var trades Trades
	err = json.Unmarshal(item, &trades)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ret := trades.ToDomain()
	return ret, nil
}

// DeleteTrades implements coinrepository.CoinRepository.
func (r *Repository) DeleteTrades(_ context.Context, id domain.CoinID) error {
	key := NewTrades(id, time.Time{}, nil)
	err := r.kv.Delete(key.Key())
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// CreateBannedCoin implements coinrepository.CoinRepository.
func (r *Repository) CreateBannedCoin(_ context.Context, bannedCoin *domain.BannedCoin) (*domain.BannedCoin, error) {
	coin := NewBannedCoin(bannedCoin)
	err := r.kv.Create(coin.Key(), coin.Value())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return bannedCoin, nil
}

// ListBannedCoins implements coinrepository.CoinRepository.
func (r *Repository) ListBannedCoins(_ context.Context) ([]*domain.BannedCoin, error) {
	var coins []*BannedCoin
	items, err := r.kv.List([]byte(bannedCoinPrefix))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, item := range items {
		var coin BannedCoin
		err := json.Unmarshal(item, &coin)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		coins = append(coins, &coin)
	}
	var ret []*domain.BannedCoin
	for _, coin := range coins {
		ret = append(ret, coin.ToDomain())
	}
	return ret, nil
}

// GetBannedCoin implements coinrepository.CoinRepository.
func (r *Repository) GetBannedCoin(_ context.Context, coinID domain.CoinID) (*domain.BannedCoin, error) {
	key := BannedCoinKey(coinID)
	item, err := r.kv.Get(key)
	if err != nil {
		if errors.Is(err, keyvalue.ErrKeyNotFound) {
			return nil, coinrepository.ErrBannedCoinNotFound
		}
		return nil, errors.WithStack(err)
	}
	var ret BannedCoin
	err = json.Unmarshal(item, &ret)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret.ToDomain(), nil
}

// DeleteBannedCoin implements coinrepository.CoinRepository.
func (r *Repository) DeleteBannedCoin(_ context.Context, bannedCoin *domain.BannedCoin) error {
	err := r.kv.Delete(BannedCoinKey(bannedCoin.CoinID()))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
