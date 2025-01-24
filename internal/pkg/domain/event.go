package domain

type Event interface {
	Topic() string
	Payload() []byte
}
