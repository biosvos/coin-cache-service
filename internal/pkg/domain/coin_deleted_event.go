package domain

import (
	"encoding/json"
	"time"
)

var _ Event = (*CoinDeletedEvent)(nil)

const CoinDeletedEventTopic = "coin.deleted"

type CoinDeletedEvent struct {
	CoinID    CoinID `json:"coin_id"`
	DeletedAt time.Time
}

func NewCoinDeletedEvent(deletedAt time.Time, coinID CoinID) *CoinDeletedEvent {
	return &CoinDeletedEvent{DeletedAt: deletedAt, CoinID: coinID}
}

func ParseCoinDeletedEvent(payload []byte) *CoinDeletedEvent {
	var event CoinDeletedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func (e *CoinDeletedEvent) Topic() string {
	return CoinDeletedEventTopic
}

func (e *CoinDeletedEvent) Payload() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return payload
}
