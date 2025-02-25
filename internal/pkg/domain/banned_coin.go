package domain

import "time"

type BannedCoin struct {
	coinID   CoinID
	bannedAt time.Time
	period   time.Duration
}

func NewBannedCoin(coinID CoinID, bannedAt time.Time, period time.Duration) *BannedCoin {
	return &BannedCoin{coinID: coinID, bannedAt: bannedAt, period: period}
}

func (b *BannedCoin) IsBanOver(now time.Time) bool {
	return b.bannedAt.Add(b.period).After(now)
}

func (b *BannedCoin) CoinID() CoinID {
	return b.coinID
}

func (b *BannedCoin) BannedAt() time.Time {
	return b.bannedAt
}

func (b *BannedCoin) ExpiredAt() time.Time {
	return b.bannedAt.Add(b.period)
}

func (b *BannedCoin) Period() time.Duration {
	return b.period
}
