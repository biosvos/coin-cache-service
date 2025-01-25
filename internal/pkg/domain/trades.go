package domain

import "time"

type Trades struct {
	coinID     CoinID
	modifiedAt time.Time
	trades     []*Trade
}

func NewTrades(coinID CoinID, modifiedAt time.Time, trades []*Trade) *Trades {
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
