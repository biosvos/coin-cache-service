package real

import (
	"encoding/json"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type BannedCoin struct {
	ID       string
	BannedAt time.Time
	Period   time.Duration
}

func NewBannedCoin(coin *domain.BannedCoin) *BannedCoin {
	return &BannedCoin{
		ID:       string(coin.CoinID()),
		BannedAt: coin.BannedAt(),
		Period:   coin.Period(),
	}
}

const bannedCoinPrefix = "banned_coin:"

func (c *BannedCoin) Key() []byte {
	return []byte(bannedCoinPrefix + c.ID)
}

func (c *BannedCoin) Value() []byte {
	bytes, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (c *BannedCoin) ToDomain() *domain.BannedCoin {
	return domain.NewBannedCoin(domain.CoinID(c.ID), c.BannedAt, c.Period)
}
