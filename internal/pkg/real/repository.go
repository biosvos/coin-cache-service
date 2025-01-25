package real

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/biosvos/coin-cache-service/internal/app/coinrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	badger "github.com/dgraph-io/badger/v4"
)

var _ coinrepository.CoinRepository = (*Repository)(nil)

type Repository struct {
	db *badger.DB
}

func NewRepository(path string) *Repository {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		log.Fatal(err)
	}
	return &Repository{db: db}
}

func (r *Repository) Close() {
	r.db.Close()
}

// CreateCoin implements coinrepository.CoinRepository.
func (r *Repository) CreateCoin(ctx context.Context, domainCoin *domain.Coin) (*domain.Coin, error) {
	err := r.db.Update(func(txn *badger.Txn) error {
		coin := NewCoin(domainCoin)
		_, err := txn.Get(coin.Key())
		if err == nil {
			return errors.New("coin already exists")
		}
		return txn.Set(coin.Key(), coin.Value())
	})
	if err != nil {
		return nil, err
	}
	return domainCoin, nil
}

// ListCoins implements coinrepository.CoinRepository.
func (r *Repository) ListCoins(ctx context.Context) ([]*domain.Coin, error) {
	var coins []*Coin
	err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		prefix := []byte(coinPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var coin Coin
				err := json.Unmarshal(v, &coin)
				if err != nil {
					return err
				}
				coins = append(coins, &coin)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var ret []*domain.Coin
	for _, coin := range coins {
		ret = append(ret, coin.ToDomain())
	}
	return ret, nil
}

// UpdateCoin implements coinrepository.CoinRepository.
func (r *Repository) UpdateCoin(ctx context.Context, domainCoin *domain.Coin) (*domain.Coin, error) {
	err := r.db.Update(func(txn *badger.Txn) error {
		coin := NewCoin(domainCoin)
		_, err := txn.Get(coin.Key())
		if err != nil {
			return err
		}
		return txn.Set(coin.Key(), coin.Value())
	})
	if err != nil {
		return nil, err
	}
	return domainCoin, nil
}

// DeleteCoin implements coinrepository.CoinRepository.
func (r *Repository) DeleteCoin(ctx context.Context, domainCoin *domain.Coin) error {
	return r.db.Update(func(txn *badger.Txn) error {
		coin := NewCoin(domainCoin)
		return txn.Delete(coin.Key())
	})
}

// SaveTrades implements coinrepository.CoinRepository.
func (r *Repository) SaveTrades(ctx context.Context, id domain.CoinID, domainTrades []*domain.Trade) error {
	return r.db.Update(func(txn *badger.Txn) error {
		trades := NewTrades(id, domainTrades)
		return txn.Set(trades.Key(), trades.Value())
	})
}

// ListTrades implements coinrepository.CoinRepository.
func (r *Repository) ListTrades(ctx context.Context, id domain.CoinID) ([]*domain.Trade, error) {
	var domainTrades []*domain.Trade
	err := r.db.View(func(txn *badger.Txn) error {
		trades := NewTrades(id, nil)
		item, err := txn.Get(trades.Key())
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			var trades Trades
			err := json.Unmarshal(v, &trades)
			if err != nil {
				return err
			}
			domainTrades = trades.ToDomain()
			return nil
		})
	})
	return domainTrades, err
}

// DeleteTrades implements coinrepository.CoinRepository.
func (r *Repository) DeleteTrades(ctx context.Context, id domain.CoinID) error {
	return r.db.Update(func(txn *badger.Txn) error {
		trades := NewTrades(id, nil)
		return txn.Delete(trades.Key())
	})
}

// CreateBannedCoin implements coinrepository.CoinRepository.
func (r *Repository) CreateBannedCoin(ctx context.Context, bannedCoin *domain.BannedCoin) (*domain.BannedCoin, error) {
	panic("unimplemented")
}

// ListBannedCoins implements coinrepository.CoinRepository.
func (r *Repository) ListBannedCoins(ctx context.Context) ([]*domain.BannedCoin, error) {
	var coins []*BannedCoin
	err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		prefix := []byte(bannedCoinPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var coin BannedCoin
				err := json.Unmarshal(v, &coin)
				if err != nil {
					return err
				}
				coins = append(coins, &coin)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var ret []*domain.BannedCoin
	for _, coin := range coins {
		ret = append(ret, coin.ToDomain())
	}
	return ret, nil
}

// UpdateBannedCoin implements coinrepository.CoinRepository.
func (r *Repository) UpdateBannedCoin(ctx context.Context, bannedCoin *domain.BannedCoin) (*domain.BannedCoin, error) {
	panic("unimplemented")
}

// DeleteBannedCoin implements coinrepository.CoinRepository.
func (r *Repository) DeleteBannedCoin(ctx context.Context, bannedCoin *domain.BannedCoin) error {
	panic("unimplemented")
}
