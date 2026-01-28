package kafkax

import (
	"context"
	"time"
)

type Handler func(ctx context.Context, msg *Message) error

// Message represents a Kafka message to be sent
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Partition int
	Offset    int64
	Time      time.Time
}
