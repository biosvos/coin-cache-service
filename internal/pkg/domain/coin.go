package domain

import "time"

type CoinID string

type Coin struct {
	id         CoinID
	caution    bool
	modifiedAt time.Time
}

func (c *Coin) SetModifiedAt(now time.Time) *Coin {
	return NewCoin(c.id, c.caution, now)
}

func NewCoin(id CoinID, caution bool, modifiedAt time.Time) *Coin {
	return &Coin{id: id, caution: caution, modifiedAt: modifiedAt}
}

func (c *Coin) ID() CoinID {
	return c.id
}

func (c *Coin) IsOld(now time.Time) bool {
	const coinPeriod = time.Minute * 10
	return c.modifiedAt.Add(coinPeriod).Before(now)
}

func (c *Coin) IsDanger() bool {
	return c.caution
}

func (c *Coin) ModifiedAt() time.Time {
	return c.modifiedAt
}
