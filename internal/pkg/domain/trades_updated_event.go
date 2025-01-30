package domain

import "encoding/json"

var _ Event = (*TradesUpdatedEvent)(nil)

const TradesUpdatedEventTopic = "trades.updated"

type TradesUpdatedEvent struct {
	CoinID CoinID `json:"coin_id"`
}

func NewTradesUpdatedEvent(coinID CoinID) *TradesUpdatedEvent {
	return &TradesUpdatedEvent{CoinID: coinID}
}

func ParseTradesUpdatedEvent(payload []byte) *TradesUpdatedEvent {
	var event TradesUpdatedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func (e *TradesUpdatedEvent) Topic() string {
	return TradesUpdatedEventTopic
}

func (e *TradesUpdatedEvent) Payload() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return payload
}
