package domain

import "encoding/json"

var _ Event = (*BannedCoinCreatedEvent)(nil)

const BannedCoinCreatedEventTopic = "banned_coin.created"

type BannedCoinCreatedEvent struct {
	CoinID CoinID `json:"coin_id"`
}

func ParseBannedCoinCreatedEvent(payload []byte) *BannedCoinCreatedEvent {
	var event BannedCoinCreatedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func NewBannedCoinCreatedEvent(coinID CoinID) *BannedCoinCreatedEvent {
	return &BannedCoinCreatedEvent{CoinID: coinID}
}

func (e *BannedCoinCreatedEvent) Topic() string {
	return BannedCoinCreatedEventTopic
}

func (e *BannedCoinCreatedEvent) Payload() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return payload
}
