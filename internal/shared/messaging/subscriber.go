package messaging

import (
	"context"
	"strings"
)

type Topic string

func (t Topic) IsFifo() bool {
	return strings.HasSuffix(string(t), ".fifo")
}

type BackMessage interface {
	GetPayload() Payload
	Nack() bool
	Ack() bool
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic Topic) (<-chan BackMessage, error)
	Close() error
}
