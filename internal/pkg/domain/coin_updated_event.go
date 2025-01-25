package domain

import (
	"encoding/json"
	"time"
)

var _ Event = (*CoinUpdatedEvent)(nil)

const CoinUpdatedEventTopic = "coin.updated"

type CoinUpdatedEvent struct {
	CoinID    CoinID    `json:"coin_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ParseCoinUpdatedEvent(payload []byte) *CoinUpdatedEvent {
	var event CoinUpdatedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func NewCoinUpdatedEvent(updatedAt time.Time, coinID CoinID) *CoinUpdatedEvent {
	return &CoinUpdatedEvent{UpdatedAt: updatedAt, CoinID: coinID}
}

func (e *CoinUpdatedEvent) Topic() string {
	return CoinUpdatedEventTopic
}

func (e *CoinUpdatedEvent) Payload() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return payload
}
