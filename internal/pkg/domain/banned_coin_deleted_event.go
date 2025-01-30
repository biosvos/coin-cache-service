package domain

import "encoding/json"

var _ Event = (*BannedCoinDeletedEvent)(nil)

const BannedCoinDeletedEventTopic = "banned_coin.deleted"

type BannedCoinDeletedEvent struct {
	CoinID CoinID `json:"coin_id"`
}

func NewBannedCoinDeletedEvent(coinID CoinID) *BannedCoinDeletedEvent {
	return &BannedCoinDeletedEvent{CoinID: coinID}
}

func ParseBannedCoinDeletedEvent(payload []byte) *BannedCoinDeletedEvent {
	var event BannedCoinDeletedEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		panic(err)
	}
	return &event
}

func (e *BannedCoinDeletedEvent) Topic() string {
	return BannedCoinDeletedEventTopic
}

func (e *BannedCoinDeletedEvent) Payload() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return payload
}
