package realrepository

import (
	"encoding/json"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type BannedCoin struct {
	ID       string        `json:"id,omitempty"`
	BannedAt time.Time     `json:"banned_at,omitempty"`
	Period   time.Duration `json:"period,omitempty"`
}

func NewBannedCoin(coin *domain.BannedCoin) *BannedCoin {
	return &BannedCoin{
		ID:       string(coin.CoinID()),
		BannedAt: coin.BannedAt(),
		Period:   coin.Period(),
	}
}

const bannedCoinPrefix = "banned_coin:"

func BannedCoinKey(coinID domain.CoinID) []byte {
	return []byte(bannedCoinPrefix + string(coinID))
}

func (c *BannedCoin) Key() []byte {
	return BannedCoinKey(domain.CoinID(c.ID))
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
