package realrepository

import (
	"encoding/json"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type Trades struct {
	CoinID     domain.CoinID `json:"coin_id,omitempty"`
	ModifiedAt time.Time     `json:"modified_at,omitempty"`
	Trades     []*Trade      `json:"trades,omitempty"`
}

func NewTrades(coinID domain.CoinID, modifiedAt time.Time, domainTrades []*domain.Trade) *Trades {
	var trades []*Trade
	for _, trade := range domainTrades {
		trades = append(trades, NewTrade(trade))
	}
	return &Trades{CoinID: coinID, ModifiedAt: modifiedAt, Trades: trades}
}

type Trade struct {
	Date         time.Time `json:"date,omitempty"`
	LastPrice    string    `json:"last_price,omitempty"`
	OpeningPrice string    `json:"opening_price,omitempty"`
	MaxPrice     string    `json:"max_price,omitempty"`
	MinPrice     string    `json:"min_price,omitempty"`
}

func (t *Trade) ToDomain() *domain.Trade {
	return domain.NewTrade(
		t.Date,
		domain.Price(t.LastPrice),
		domain.Price(t.OpeningPrice),
		domain.Price(t.MaxPrice),
		domain.Price(t.MinPrice),
	)
}

func NewTrade(domainTrade *domain.Trade) *Trade {
	return &Trade{
		Date:         domainTrade.Date(),
		LastPrice:    string(domainTrade.LastPrice()),
		OpeningPrice: string(domainTrade.OpeningPrice()),
		MaxPrice:     string(domainTrade.MaxPrice()),
		MinPrice:     string(domainTrade.MinPrice()),
	}
}

const tradesPrefix = "trades:"

func (t *Trades) Key() []byte {
	return []byte(tradesPrefix + string(t.CoinID))
}

func (t *Trades) Value() []byte {
	bytes, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (t *Trades) ToDomain() *domain.Trades {
	var trades []*domain.Trade
	for _, trade := range t.Trades {
		trades = append(trades, trade.ToDomain())
	}
	return domain.NewTrades(t.CoinID, t.ModifiedAt, trades)
}
