package domain

import (
	"encoding/json"
)

var _ Event = (*CoinBannedEvent)(nil)

type CoinBannedEvent struct {
	CoinID CoinID `json:"coin_id"`
}

func ParseCoinBannedEvent(payload []byte) *CoinBannedEvent {
	var event CoinBannedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func NewCoinBannedEvent(coinID CoinID) *CoinBannedEvent {
	return &CoinBannedEvent{CoinID: coinID}
}

func (c *CoinBannedEvent) Topic() string {
	return "coin.banned"
}

func (c *CoinBannedEvent) Payload() []byte {
	return []byte(c.CoinID)
}
