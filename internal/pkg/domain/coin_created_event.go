package domain

import (
	"encoding/json"
	"time"
)

var _ Event = (*CoinCreatedEvent)(nil)

const CoinCreatedEventTopic = "coin.created"

type CoinCreatedEvent struct {
	CoinID    CoinID    `json:"coin_id"`
	CreatedAt time.Time `json:"created_at"`
}

func ParseCoinCreatedEvent(payload []byte) *CoinCreatedEvent {
	var event CoinCreatedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func NewCoinCreatedEvent(createdAt time.Time, coinID CoinID) *CoinCreatedEvent {
	return &CoinCreatedEvent{CreatedAt: createdAt, CoinID: coinID}
}

func (e *CoinCreatedEvent) Topic() string {
	return CoinCreatedEventTopic
}

func (e *CoinCreatedEvent) Payload() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return payload
}
