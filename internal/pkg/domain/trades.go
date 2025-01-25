package domain

import (
	"sort"
	"time"
)

type Trades struct {
	coinID     CoinID
	modifiedAt time.Time
	trades     []*Trade
}

func NewTrades(coinID CoinID, modifiedAt time.Time, trades []*Trade) *Trades {
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].Date().Before(trades[j].Date())
	})
	return &Trades{coinID: coinID, modifiedAt: modifiedAt, trades: trades}
}

func (t *Trades) CoinID() CoinID {
	return t.coinID
}

func (t *Trades) ModifiedAt() time.Time {
	return t.modifiedAt
}

func (t *Trades) Trades() []*Trade {
	return t.trades
}

func (t *Trades) Size() int {
	return len(t.trades)
}

func (t *Trades) IsEnoughTrade() bool {
	const enoughTrade = 20
	return len(t.trades) >= enoughTrade
}

func (t *Trades) LastPrice() Price {
	return t.trades[len(t.trades)-1].LastPrice()
}
