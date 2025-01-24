package real

import (
	"encoding/json"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type Coin struct {
	ID         string
	Danger     bool
	ModifiedAt time.Time
}

const coinPrefix = "coin:"

func NewCoin(domainCoin *domain.Coin) *Coin {
	return &Coin{
		ID:         string(domainCoin.ID()),
		Danger:     domainCoin.IsDanger(),
		ModifiedAt: domainCoin.ModifiedAt(),
	}
}

func (c *Coin) Key() []byte {
	return []byte(coinPrefix + c.ID)
}

func (c *Coin) Value() []byte {
	bytes, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (c *Coin) ToDomain() *domain.Coin {
	return domain.NewCoin(
		domain.CoinID(c.ID),
		c.Danger,
		c.ModifiedAt,
	)
}
