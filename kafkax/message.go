package kafkax

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Handler func(ctx context.Context, msg *ConsumedMessage) error

// Header represents a Kafka message header
type Header struct {
	Key   string
	Value []byte
}

// Message represents a Kafka message to be sent
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Partition int
	Offset    int64
	Time      time.Time
	Headers   []Header
}

// ConsumedMessage represents a message received from Kafka with commit capability
type ConsumedMessage struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Time      time.Time

	kafkaMsg kafka.Message
	reader   *kafka.Reader
}

// Commit commits the consumed message offset
func (m *ConsumedMessage) Commit(ctx context.Context) error {
	if m == nil || m.reader == nil {
		return nil
	}
	return m.reader.CommitMessages(ctx, m.kafkaMsg)
}
