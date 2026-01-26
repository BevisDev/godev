package kafkax

import "context"

type Handler func(ctx context.Context, msg *Message) error

type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Partition int
	Offset    int64
}
