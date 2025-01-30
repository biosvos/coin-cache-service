package domain

import "encoding/json"

var _ Event = (*TradesDeletedEvent)(nil)

const TradesDeletedEventTopic = "trades.deleted"

type TradesDeletedEvent struct {
	CoinID CoinID `json:"coin_id"`
}

func NewTradesDeletedEvent(coinID CoinID) *TradesDeletedEvent {
	return &TradesDeletedEvent{CoinID: coinID}
}

func ParseTradesDeletedEvent(payload []byte) *TradesDeletedEvent {
	var event TradesDeletedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func (e *TradesDeletedEvent) Topic() string {
	return TradesDeletedEventTopic
}

func (e *TradesDeletedEvent) Payload() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return payload
}
